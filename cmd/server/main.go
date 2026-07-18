package main

import "fmt"

type Speaker interface {
	Speak() string
}

type Dog struct {
	Name string
}

func (d *Dog) Speak() string {
	return d.Name + "dit Wouf!"
}

func (d *Dog) Age() int {
	return 3
}

type Cat struct {
	Name string
}

func (c Cat) Speak() string {
	return c.Name + "dit Miaou!"
}

type Bird struct {
	Name string
}

func (b Bird) Speak() string {
	return b.Name + "dit cuicui!"
}

func main() {
	animals := []Speaker{
		&Dog{Name: "Rex"},
		Cat{Name: "Felix"},
		Bird{Name: "TIti"},
	}

	for _, a := range animals {
		fmt.Println(a.Speak())
		if d, ok := a.(*Dog); ok {
			fmt.Println("Age:", d.Age())
		}

	}
}
