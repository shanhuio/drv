package jarvis

import (
	"shanhu.io/drv/drvapi"
	"shanhu.io/std/errcode"
)

type appClosure struct {
	floors   [][]*drvapi.AppMeta        // floor layout of app deps
	floorMap map[string]int             // map from app name to floor
	apps     map[string]*drvapi.AppMeta // all apps, indexed by name

	// trace tracks circle dependencies.
	trace    []string
	traceMap map[string]bool
}

func newAppClosure() *appClosure {
	return &appClosure{
		apps:     make(map[string]*drvapi.AppMeta),
		floorMap: make(map[string]int),
		traceMap: make(map[string]bool),
	}
}

func (c *appClosure) add(
	q appQuerier, name string, ignore map[string]bool,
) (int, error) {
	if ignore != nil && ignore[name] {
		return 0, nil
	}

	if c.traceMap[name] {
		var circle []string
		for i, t := range c.trace {
			if t == name {
				circle = c.trace[i:]
				break
			}
		}
		return 0, errcode.Internalf(
			"app %q has circle dependency: %q",
			name, circle,
		)
	}
	c.traceMap[name] = true
	c.trace = append(c.trace, name)
	defer func() {
		delete(c.traceMap, name)
		c.trace = c.trace[:len(c.trace)-1]
	}()

	f, ok := c.floorMap[name]
	if ok {
		return f, nil // already in the closure
	}

	m, err := q.meta(name)
	if err != nil {
		return 0, err
	}
	c.apps[name] = m

	floor := 0
	for _, dep := range m.Deps {
		f, err := c.add(q, dep, ignore)
		if err != nil {
			return 0, errcode.Annotatef(err, "import %q", dep)
		}
		if floor <= f {
			floor = f + 1
		}
	}

	// put on the proper floor
	for len(c.floors) <= floor {
		c.floors = append(c.floors, nil)
	}
	c.floors[floor] = append(c.floors[floor], m)
	c.floorMap[name] = floor
	return floor, nil
}

func (c *appClosure) appSet() map[string]bool {
	m := make(map[string]bool)
	for name := range c.apps {
		m[name] = true
	}
	return m
}

func (c *appClosure) Floors() [][]*drvapi.AppMeta { return c.floors }

func (c *appClosure) RevFloors() [][]*drvapi.AppMeta {
	rev := make([][]*drvapi.AppMeta, len(c.floors))
	n := len(c.floors)
	for i, f := range c.floors {
		rev[n-1-i] = f
	}
	return rev
}
