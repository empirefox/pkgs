package pkgs

import "go/types"

type TagPath struct {
	Path  []string
	Value string
}

func findStruct(p *Package, st *types.Struct) (*Struct, bool) {
	for _, s := range p.StructTypes {
		if st == s.Underline {
			return s, true
		}
	}
	return nil, false
}

func (s *Struct) ComputePkgTagPaths(tag string) []TagPath {
	return s.computePkgTagPaths(nil, tag)
}

func (s *Struct) computePkgTagPaths(parent []string, tag string) (ps []TagPath) {
	for _, field := range s.IntuitiveFields {
		if v := field.Tag.Get(tag); v != "" {
			ps = append(ps, TagPath{
				Path:  append(parent, field.Name),
				Value: v,
			})
		} else if ul, ok := field.UnderlineType().(*types.Struct); ok {
			if sub, ok := findStruct(s.Pkg, ul); ok {
				subs := sub.computePkgTagPaths(append(parent, field.Name), tag)
				ps = append(ps, subs...)
			}
		}
	}
	return
}
