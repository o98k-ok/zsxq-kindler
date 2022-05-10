package core

import "fmt"

type Menu interface {
	Name() string
	Link() string
}

type PresetMenu struct {
	gid   uint64
	name  string
	pType string
	count int
}

func (p *PresetMenu) Name() string {
	return p.name
}

func (p *PresetMenu) Link() string {
	return fmt.Sprintf("/menus/preset/%s", p.pType)
}

type CustomMenu struct {
	gid   uint64
	name  string
	hId   int64
	count int
}

func (c *CustomMenu) Name() string {
	return c.name
}

func (c *CustomMenu) Link() string {
	return fmt.Sprintf("/menus/custom/%d", c.hId)
}
