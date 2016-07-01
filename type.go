// +build go1.5

package pkgs

import (
	"go/types"
	"reflect"
	"strings"
)

type Field struct {
	Name          string
	Anonymous     bool
	Exported      bool
	TypeString    string
	IsPtr         bool
	Tag           reflect.StructTag
	underlineType types.Type
}

func NewField(typesVar *types.Var, tag string, p *Package) *Field {
	field := &Field{
		Name:          typesVar.Name(),
		Anonymous:     typesVar.Anonymous(),
		Exported:      typesVar.Exported(),
		Tag:           reflect.StructTag(tag),
		underlineType: typesVar.Type().Underlying(),
	}
	ptr, ok := field.underlineType.Underlying().(*types.Pointer)
	if ok {
		field.IsPtr = ok
		field.underlineType = ptr.Elem().Underlying()
	}
	// TODO compute the type?
	field.TypeString = strings.Replace(typesVar.Type().String(), p.Dir+".", "", -1)
	return field
}

func (field *Field) UnderlineType() types.Type {
	return field.underlineType
}

// All are simpe types

type Basic struct {
	Name string
	Type string
	Doc  string
}

func NewBasic(name string, t *types.Basic) *Basic {
	return &Basic{Name: name, Type: t.Name()}
}

type Struct struct {
	Name string
	Doc  string

	// all literal fields
	Fields   []*Field
	FieldMap map[string]*Field

	// all fact basic fields, exclude literal Anonymous
	// fields from Anonymous are not equal with Anonymous struct fields
	IntuitiveFields   []*Field
	IntuitiveFieldMap map[string]*Field

	Underline *types.Struct

	Pkg *Package
}

func NewStruct(name string, t *types.Struct, p *Package) *Struct {
	s := &Struct{
		Name:              name,
		FieldMap:          make(map[string]*Field),
		IntuitiveFieldMap: make(map[string]*Field),
		Underline:         t,
		Pkg:               p,
	}
	s.buildFields(t, p)
	return s
}

// Only this package scope element
type Array struct {
	Name     string
	Elem     string
	IsStruct bool
	IsPtr    bool
	Doc      string
}

// NewArray elemTyp is element type
func NewArray(name string, elemTyp types.Type, scope *types.Scope) *Array {
	arr := &Array{Name: name}

	switch u := elemTyp.Underlying().(type) {
	case *types.Basic:
		arr.Elem = u.Name()

	case *types.Struct:
		arr.Elem, arr.IsStruct = GetScopeStructType(u, scope)
		if !arr.IsStruct {
			return nil
		}

	case *types.Pointer:
		if arr = NewArray(name, u.Elem(), scope); arr != nil {
			arr.IsPtr = true
		}

	default:
		log.WithField(name, u.String()).Infoln("Element type not supported")
		return nil
	}

	return arr
}

// Only this package scope element
type Map struct {
	Array
	Key string
}

// NewMap elemTyp is element type
func NewMap(name string, keyTyp, elemTyp types.Type, scope *types.Scope) *Map {
	kt, ok := keyTyp.(*types.Basic)
	if !ok {
		return nil
	}
	arr := NewArray(name, elemTyp, scope)
	if arr == nil {
		return nil
	}
	return &Map{Array: *arr, Key: kt.Name()}
}

// GetScopeStructType find the struct name from scope
func GetScopeStructType(et *types.Struct, scope *types.Scope) (string, bool) {
	for _, n := range scope.Names() {
		if obj, ok := scope.Lookup(n).(*types.TypeName); ok {
			typ, ok := obj.Type().(*types.Named).Underlying().(*types.Struct)
			if ok && typ == et {
				return n, true
			}
		}
	}
	return "", false
}

func (s *Struct) buildFields(t *types.Struct, p *Package) {
	var (
		ts       = []*types.Struct{t}
		fields   []*Field
		fieldMap = make(map[string]*Field)
		conflict = make(map[string]bool)
	)
	for len(ts) > 0 {
		ts = buildFields(ts, &fields, fieldMap, conflict, p)
	}

	i := 0
	for _, field := range fields {
		// add to Fields
		if i < t.NumFields() {
			s.Fields = append(s.Fields, field)
			s.FieldMap[field.Name] = field
			i++
		}
		// add to IntuitiveFields
		if field.Anonymous {
			_, ok := field.UnderlineType().(*types.Struct)
			if ok {
				continue
			}
		}
		s.IntuitiveFields = append(s.IntuitiveFields, field)
		s.IntuitiveFieldMap[field.Name] = field
	}
}

// buildFields BFS
func buildFields(ts []*types.Struct,
	fields *[]*Field, // big pool
	fieldMap map[string]*Field,
	conflict map[string]bool,
	p *Package) (nextDeep []*types.Struct) {

	var levelFields []*Field // small poll
	var levelFieldMap = make(map[string]*Field)

	// add all literal fields of this level
	for _, t := range ts {
		// add all literal fields of one struct
		for i := 0; i < t.NumFields(); i++ {
			field := NewField(t.Field(i), t.Tag(i), p)
			name := field.Name
			if conflict[name] {
				continue
			}
			if _, ok := fieldMap[name]; ok {
				continue
			}
			_, ok := levelFieldMap[name]
			if ok {
				conflict[name] = true
				continue
			}

			// add to small pool
			levelFields = append(levelFields, field)
			levelFieldMap[name] = field
		}
	}

	// add anon fields
	for _, field := range levelFields {
		if ok := conflict[field.Name]; !ok {
			// filter to big pool
			*fields = append(*fields, field)
			fieldMap[field.Name] = field

			// filter nextDeep structs
			if field.Anonymous {
				anon, ok := field.UnderlineType().(*types.Struct)
				if ok {
					nextDeep = append(nextDeep, anon)
				}
			}
		}
	}

	return
}
