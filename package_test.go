package pkgs

import (
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPackage(t *testing.T) {

	Convey("Parse fixture foo package", t, func() {
		files, err := filepath.Glob("./fixture/foo/*.go")
		So(err, ShouldBeNil)

		pkg := NewPackage(files...)
		So(pkg.Dir, ShouldEqual, ".")
		So(pkg.Name, ShouldEqual, "foo")

		bobTyp := pkg.StructTypes["Bob"]
		So(bobTyp, ShouldNotBeNil)
		So(bobTyp.Name, ShouldEqual, "Bob")
		So(strings.TrimSpace(bobTyp.Doc), ShouldEqual, "bob, doc")
		So(len(bobTyp.Fields), ShouldEqual, 3)
		So(len(bobTyp.IntuitiveFields), ShouldEqual, 5)

		So(bobTyp.IntuitiveFieldMap["ID"], ShouldBeNil)

		nameField := bobTyp.FieldMap["Name"]
		So(nameField, ShouldNotBeNil)
		So(nameField.Tag, ShouldEqual, `MGR:";lmax(16)"`)

		barField := bobTyp.IntuitiveFieldMap["Bar"]
		So(barField, ShouldNotBeNil)
		So(barField.Tag, ShouldEqual, `VIEW:";lmax(16)"`)
		So(barField.TypeString, ShouldEqual, "string")

		fooField := bobTyp.FieldMap["Foo"]
		So(fooField, ShouldNotBeNil)
		So(fooField.TypeString, ShouldEqual, "*Foo")

		aliceType := pkg.StructTypes["Alice"]
		So(aliceType, ShouldNotBeNil)
		ps := aliceType.ComputePkgTagPaths("VIEW")
		psResult := []TagPath{
			{
				Path:  []string{"Name"},
				Value: ";lmin(16)",
			},
			{
				Path:  []string{"Foo", "Bar"},
				Value: ";lmax(16)",
			},
		}
		So(ps, ShouldResemble, psResult)
	})
}
