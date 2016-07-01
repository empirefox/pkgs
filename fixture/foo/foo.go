// +build ignore

package foo

import "github.com/jinzhu/gorm"

//type Model struct {
//	ID        uint `gorm:"primary_key"`
//	CreatedAt time.Time
//	UpdatedAt time.Time
//	DeletedAt *time.Time `sql:"index"`
//}

type Foo struct {
	ID  uint
	Bar string `VIEW:";lmax(16)"`
}

type Alice struct {
	ID   uint
	Name string `VIEW:";lmin(16)"`
	Foo  *Foo
}

// bob, doc
type Bob struct {
	*Foo
	gorm.Model
	Name string `MGR:";lmax(16)"`
}

type Boys []Bob

type Boyss []*Bob

type Int int

type Ints []int

type IntMap map[int]interface{}

type BobMap map[string]Bob

type BobsMap map[int]*Bob
