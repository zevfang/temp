package pro_test

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"
	"strconv"
	"bytes"
	"sort"
	"go/format"
	"io/ioutil"
	"bufio"
	"encoding/json"

	"github.com/fatih/structtag"
	"github.com/fatih/camelcase"
	"golang.org/x/tools/go/buildutil"
)

type output struct {
	Start  int      `json:"start"`
	End    int      `json:"end"`
	Lines  []string `json:"lines"`
	Errors []string `json:"errors,omitempty"`
}

type structType struct {
	name string
	node *ast.StructType
}

type config struct {
	// first section - input & output
	file     string
	modified io.Reader
	output   string
	write    bool

	// second section - struct selection
	offset     int
	structName string
	line       string
	start, end int
	fset       *token.FileSet
	// third section - struct modification
	remove    []string
	add       []string
	override  bool
	transform string
	sort      bool
	clear     bool

	addOptions    []string
	removeOptions []string
	clearOption   bool
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain() error {
	var (
		flagFile     = flag.String("file", "", "Filename to be parsed")
		flagModified = flag.Bool("modified", false, "read an archive of modified files from standard input")
		flagOutput   = flag.String("format", "source", "Output format.By default it's the whole file. Options: [source, json]")
		flagWrite    = flag.Bool("w", false, "Write result to (source) file instead of stdout")

		// processing modes
		flagOffset = flag.Int("offset", 0, "Byte offset of the cursor position inside a struct.Can be anwhere from the comment until closing bracket")
		flagLine   = flag.String("line", "", "Line number of the field or a range of line. i.e: 4 or 4,8")
		flagStruct = flag.String("struct", "", "Struct name to be processed")

		// tag flags
		flagRemoveTags = flag.String("remove-tags", "", "Remove tags for the comma separated list of keys")
		flagClearTags  = flag.Bool("clear-tags", false, "Clear all tags")
		flagAddTags    = flag.String("add-tags", "", "Adds tags for the comma separated list of keys.Keys can contain a static value, i,e: json:foo")
		flagOverride   = flag.Bool("override", false, "Override current tags when adding tags")
		flagTransform  = flag.String("transform", "snakecase", "Transform adds a transform rule when adding tags. Current options: [snakecase, camelcase, lispcase]")
		flagSort       = flag.Bool("sort", false, "Sort sorts the tags in increasing order according to the key name")

		// option flags
		flagRemoveOptions = flag.String("remove-options", "", "Remove the comma separated list of options from the given keys, i.e: json=omitempty,hcl=squash")
		flagClearOptions  = flag.Bool("clear-options", false, "Clear all tag options")
		flagAddOptions    = flag.String("add-options", "", "Add the options per given key. i.e: json=omitempty,hcl=squash")
	)

	flag.Usage = func() {}
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return nil
	}
	cfg := &config{
		file:        *flagFile,
		line:        *flagLine,
		structName:  *flagStruct,
		offset:      *flagOffset,
		output:      *flagOutput,
		write:       *flagWrite,
		clear:       *flagClearTags,
		clearOption: *flagClearOptions,
		transform:   *flagTransform,
		sort:        *flagSort,
		override:    *flagOverride,
	}

	//询问用户索要文件
	if *flagModified {
		cfg.modified = os.Stdin
	}

	if *flagAddTags != "" {
		cfg.add = strings.Split(*flagAddTags, ",")
	}

	if *flagAddOptions != "" {
		cfg.addOptions = strings.Split(*flagAddOptions, ",")
	}

	if *flagRemoveTags != "" {
		cfg.remove = strings.Split(*flagRemoveTags, ",")
	}

	if *flagRemoveOptions != "" {
		cfg.removeOptions = strings.Split(*flagRemoveOptions, ",")
	}

	err := cfg.validate();
	if err != nil {
		return err
	}

	//解析
	node, err := cfg.parse()
	if err != nil {
		return err
	}

	//查找
	start, end, err := cfg.findSelection(node)
	if err != nil {
		return err
	}

	rewritenNode, errs := cfg.rewrite(node, start, end)
	if errs != nil {
		//判断类型是否一致
		if _, ok := errs.(*rewriteErrors); !ok {
			return errs
		}
	}

	out, err := cfg.format(rewritenNode, errs)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil

}

func (c *config) findSelection(node ast.Node) (int, int, error) {
	if c.line != "" {
		return c.lineSelection(node)
	} else if c.offset != 0 {
		return c.offsetSelection(node)
	} else if c.structName != "" {
		return c.structSelection(node)
	} else {
		return 0, 0, errors.New("-line,-offset or -struct is not passed")
	}

}

func (c *config) rewrite(node ast.Node, start, end int) (ast.Node, error) {
	errs := &rewriteErrors{make([]error, 0)}
	rewriteFunc := func(n ast.Node) bool {
		x, ok := n.(*ast.StructType)
		if !ok {
			return true
		}

		for _, f := range x.Fields.List {
			line := c.fset.Position(f.Pos()).Line
			if !(start <= line && line <= end) {
				continue
			}

			if f.Tag == nil {
				f.Tag = &ast.BasicLit{}
			}

			filedName := ""
			if len(f.Names) != 0 {
				filedName = f.Names[0].Name
			}

			if f.Names == nil {
				ident, ok := f.Type.(*ast.Ident)
				if !ok {
					continue
				}
				filedName = ident.Name
			}

			res, err := c.process(filedName, f.Tag.Value)
			if err != nil {
				errs.Append(fmt.Errorf("%s:%d:%d:%s",
					c.fset.Position(f.Pos()).Filename,
					c.fset.Position(f.Pos()).Line,
					c.fset.Position(f.Pos()).Column,
					err))
				continue
			}
			f.Tag.Value = res
		}
		return true
	}
	ast.Inspect(node, rewriteFunc)
	c.start = start
	c.end = end
	if len(errs.errs) == 0 {
		return node, nil
	}
	return node, nil
}

func (c *config) process(fileName, tagVal string) (string, error) {
	var tag string
	if tagVal != "" {
		var err error
		tag, err = strconv.Unquote(tagVal)
		if err != nil {
			return "", err
		}
	}

	tags, err := structtag.Parse(tag)
	if err != nil {
		return "", err
	}

	tags = c.removeTags(tags)
	tags, err = c.removeTagOptions(tags)
	if err != nil {
		return "", err
	}

	tags = c.clearTags(tags)
	tags = c.clearOptions(tags)

	tags, err = c.AddTag(fileName, tags)
	if err != nil {
		return "", err
	}

	tags, err = c.addTagOptions(tags)
	if err != nil {
		return "", err
	}

	if c.sort {
		sort.Sort(tags)
	}

	res := tags.String()
	if res != "" {
		res = quote(tags.String())
	}
	return res, nil
}

func (c *config) removeTags(tags *structtag.Tags) *structtag.Tags {
	if c.remove == nil || len(c.remove) == 0 {
		return tags
	}
	tags.Delete(c.remove...)
	return tags
}

func (c *config) clearTags(tags *structtag.Tags) *structtag.Tags {
	if !c.clear {
		return tags
	}
	tags.Delete(tags.Keys()...)
	return tags
}

func (c *config) clearOptions(tags *structtag.Tags) *structtag.Tags {
	if !c.clearOption {
		return tags
	}

	for _, t := range tags.Tags() {
		t.Options = nil
	}

	return tags
}

func (c *config) removeTagOptions(tags *structtag.Tags) (*structtag.Tags, error) {
	if c.removeOptions == nil || len(c.removeOptions) == 0 {
		return tags, nil
	}

	for _, val := range c.removeOptions {
		splitted := strings.Split(val, "=")
		if len(splitted) != 2 {
			return nil, errors.New("wrong syntax to remove an option. i.e key=option")
		}
		key := splitted[0]
		option := splitted[1]
		tags.DeleteOptions(key, option)
	}
	return tags, nil
}

func (c *config) addTagOptions(tags *structtag.Tags) (*structtag.Tags, error) {
	if c.addOptions == nil || len(c.addOptions) == 0 {
		return tags, nil
	}
	for _, val := range c.addOptions {
		splitted := strings.Split(val, "=")
		if len(splitted) != 2 {
			return tags, errors.New("wrong syntax to add an option. i.e key=option")
		}
		key := splitted[0]
		option := splitted[0]
		tags.AddOptions(key, option)
	}
	return tags, nil
}

//[snake_case, CamelCase, lisp-case]
func (c *config) AddTag(fieldName string, tags *structtag.Tags) (*structtag.Tags, error) {
	if c.add == nil || len(c.add) == 0 {
		return tags, nil
	}

	splitted := camelcase.Split(fieldName)
	name := ""
	unknown := false
	switch c.transform {
	case "snakecase": //go_lang
		var lowerSplitted []string
		for _, s := range splitted {
			lowerSplitted = append(lowerSplitted, strings.ToLower(s))
		}
		name = strings.Join(lowerSplitted, "_")
	case "lispcase": //go-lang
		var lowerSplitted []string
		for _, s := range splitted {
			lowerSplitted = append(lowerSplitted, strings.ToLower(s))
		}
		name = strings.Join(lowerSplitted, "-")
	case "camelcase": //goLang
		var titled []string
		for _, s := range splitted {
			titled = append(titled, strings.Title(s))
		}
		titled[0] = strings.ToLower(titled[0])
		name = strings.Join(titled, "")
	case "pascalcase": //GoLang
		var titiled []string
		for _, s := range splitted {
			titiled = append(titiled, strings.Title(s))
		}
		name = strings.Join(titiled, "")
	default:
		unknown = true //没有匹配项
	}

	for _, key := range c.add {
		splitted = strings.Split(key, ":")
		if len(splitted) == 2 {
			key = splitted[0]
			name = splitted[1]
		} else if (unknown) {
			return nil, fmt.Errorf("unknown transform option %q", c.transform)
		}

		tag, err := tags.Get(key)
		if err != nil {
			tag = &structtag.Tag{
				Key:  key,
				Name: name,
			}
		} else if c.override {
			tag.Name = name
		}

		if err := tags.Set(tag); err != nil {
			return nil, err
		}
	}
	return tags, nil
}

func (c *config) format(file ast.Node, rwErrs error) (string, error) {
	switch c.output {
	case "source":
		var buf bytes.Buffer
		err := format.Node(&buf, c.fset, file)
		if err != nil {
			return "", err
		}

		if c.write {
			err = ioutil.WriteFile(c.file, buf.Bytes(), 0)
			if err != nil {
				return "", err
			}
		}
		return buf.String(), nil
	case "json":
		//包含处理有损注释
		var buf bytes.Buffer
		err := format.Node(&buf, c.fset, file)
		if err != nil {
			return "", err
		}

		var lines []string
		scanner := bufio.NewScanner(bytes.NewBufferString(buf.String()))
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if c.start > len(lines) {
			return "", errors.New("line selection is invalid")
		}

		out := &output{
			Start: c.start,
			End:   c.end,
			Lines: lines[c.start-1:c.end],
		}

		if rwErrs != nil {
			if r, ok := rwErrs.(*rewriteErrors); ok {
				for _, err := range r.errs {
					out.Errors = append(out.Errors, err.Error())
				}
			}
		}

		o, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return "", err
		}
		return string(o), nil
	default:
		return "", fmt.Errorf("unknown output mode: %s", c.output)
	}
}

//行选择
func (c *config) lineSelection(file ast.Node) (int, int, error) {
	var err error
	splitted := strings.Split(c.line, ",")
	start, err := strconv.Atoi(splitted[0])
	if err != nil {
		return 0, 0, err
	}

	end := start
	if len(splitted) == 2 {
		end, err = strconv.Atoi(splitted[1])
		if err != nil {
			return 0, 0, err
		}
	}

	if start > end {
		return 0, 0, errors.New("wrong range. start line cannot be larger than end line")
	}

	return start, end, nil
}

//偏移量
func (c *config) offsetSelection(file ast.Node) (int, int, error) {
	structs := collectStructs(file)
	var encStruct *ast.StructType
	for _, st := range structs {
		structBegin := c.fset.Position(st.node.Pos()).Offset
		structEnd := c.fset.Position(st.node.End()).Offset

		if structBegin <= c.offset && c.offset <= structEnd {
			encStruct = st.node
			break
		}
	}
	if encStruct == nil {
		return 0, 0, errors.New("offset is not inside a struct")
	}
	start := c.fset.Position(encStruct.Pos()).Line
	end := c.fset.Position(encStruct.End()).Line
	return start, end, nil
}

//结构体名称
func (c *config) structSelection(file ast.Node) (int, int, error) {
	structs := collectStructs(file)
	var encStruct *ast.StructType
	for _, st := range structs {
		if st.name == c.structName {
			encStruct = st.node
		}
	}
	if encStruct == nil {
		return 0, 0, errors.New("struct name does not exis")
	}
	start := c.fset.Position(encStruct.Pos()).Line
	end := c.fset.Position(encStruct.End()).Line
	return start, end, nil
}

/**
  * 获取源文件中的结构体
  */
func collectStructs(node ast.Node) map[token.Pos]*structType {
	structs := make(map[token.Pos]*structType, 0)
	collectStructs := func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if t.Type == nil {
			return true
		}
		structName := t.Name.Name
		x, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}
		structs[x.Pos()] = &structType{
			name: structName,
			node: x,
		}
		return true
	}
	//Inspect 函数逐步遍历 AST 并搜索结构体
	ast.Inspect(node, collectStructs)
	return structs
}

/**
 * 解析代码
 */
func (c *config) parse() (ast.Node, error) {
	c.fset = token.NewFileSet()
	var contents interface{}
	if c.modified != nil {
		archive, err := buildutil.ParseOverlayArchive(c.modified)
		if err != nil {
			return nil, fmt.Errorf("failed to parse -modified archive: %v", err)
		}
		fc, ok := archive[c.file]
		if !ok {
			return nil, fmt.Errorf("couldn't find %s in archive", c.file)
		}
		contents = fc
	}
	return parser.ParseFile(c.fset, c.file, contents, parser.ParseComments)
}

/**
 * 输入验证
 */
func (c *config) validate() error {
	if c.file == "" {
		return errors.New("no file is passed")
	}

	if c.line == "" && c.offset == 0 && c.structName == "" {
		return errors.New("-line,-offset or -struct cannot be used together,pick one")
	}

	if c.line != "" && c.offset != 0 ||
		c.line != "" && c.structName != "" ||
		c.offset != 0 && c.structName != "" {
		return errors.New("-line, -offset or -struct cannot be used together. pick one")
	}

	if (c.add == nil || len(c.add) == 0) &&
		(c.addOptions == nil || len(c.addOptions) == 0) &&
		!c.clear &&
		!c.clearOption &&
		(c.removeOptions == nil || len(c.removeOptions) == 0) &&
		(c.remove == nil || len(c.remove) == 0) {
		return errors.New("one of " +
			"[-add-tags, -add-options, -remove-tags, -remove-options, -clear-tags, -clear-options]" +
			" should be defined")
	}

	return nil
}

func quote(tag string) string {
	return "`" + tag + "`"
}

type rewriteErrors struct {
	errs []error
}

func (r *rewriteErrors) Error() string {
	var buf bytes.Buffer
	for _, v := range r.errs {
		buf.WriteString(fmt.Sprintf("%s\n", v.Error()))
	}
	return buf.String()
}

func (r *rewriteErrors) Append(err error) {
	if err == nil {
		return
	}
	r.errs = append(r.errs, err)
}
