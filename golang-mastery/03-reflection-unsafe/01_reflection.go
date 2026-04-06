// =============================================================================
// LESSON 3.1: REFLECTION — Runtime Type Introspection
// =============================================================================
//
// reflect package lets you inspect types, read/modify values, and call
// functions at runtime. Used by encoding/json, fmt, ORMs, and DI frameworks.
//
// TWO CORE TYPES:
//   reflect.Type  — describes the Go type (immutable, comparable)
//   reflect.Value — holds a runtime value (can be read/written)
//
// PERFORMANCE: Reflection is 10-100x slower than direct code.
// Use it for configuration/setup, NOT in hot loops.
// =============================================================================

package main

import (
	"fmt"
	"reflect"
	"strings"
)

// =============================================================================
// PART 1: Type Introspection
// =============================================================================

type User struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"min=0,max=150"`
	IsAdmin  bool   `json:"is_admin"`
	password string // unexported — reflection can see but not set from outside
}

func inspectType(v interface{}) {
	t := reflect.TypeOf(v)

	// If pointer, get the underlying element type
	if t.Kind() == reflect.Ptr {
		fmt.Printf("Pointer to: %s\n", t.Elem().Name())
		t = t.Elem()
	}

	fmt.Printf("Type: %s\n", t.Name())
	fmt.Printf("Kind: %s\n", t.Kind())
	fmt.Printf("Package: %s\n", t.PkgPath())
	fmt.Printf("Size: %d bytes\n", t.Size())
	fmt.Printf("Alignment: %d bytes\n", t.Align())

	if t.Kind() == reflect.Struct {
		fmt.Printf("Fields (%d):\n", t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fmt.Printf("  [%d] %-10s %-10s exported=%-5v tag=%s\n",
				i, f.Name, f.Type, f.IsExported(), f.Tag)
		}
	}

	// Methods (includes methods on pointer receiver)
	fmt.Printf("Methods (%d):\n", t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		fmt.Printf("  [%d] %s %s\n", i, m.Name, m.Type)
	}
}

func (u User) Greet() string          { return "Hello, " + u.Name }
func (u *User) SetName(name string)   { u.Name = name }

// =============================================================================
// PART 2: Value Inspection and Modification
// =============================================================================

func inspectAndModify(v interface{}) {
	val := reflect.ValueOf(v)

	// Must be a pointer to modify
	if val.Kind() != reflect.Ptr {
		fmt.Println("ERROR: Must pass pointer to modify values")
		return
	}

	val = val.Elem() // dereference pointer
	t := val.Type()

	fmt.Printf("\n=== Modifying %s ===\n", t.Name())

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)

		fmt.Printf("Field %s: ", fieldType.Name)

		if !field.CanSet() {
			fmt.Printf("CANNOT SET (unexported)\n")
			continue
		}

		switch field.Kind() {
		case reflect.String:
			old := field.String()
			field.SetString(strings.ToUpper(old))
			fmt.Printf("%q → %q\n", old, field.String())
		case reflect.Int:
			old := field.Int()
			field.SetInt(old + 1)
			fmt.Printf("%d → %d\n", old, field.Int())
		case reflect.Bool:
			old := field.Bool()
			field.SetBool(!old)
			fmt.Printf("%v → %v\n", old, field.Bool())
		default:
			fmt.Printf("(unhandled kind: %s)\n", field.Kind())
		}
	}
}

// =============================================================================
// PART 3: Building a struct validator using reflection
// =============================================================================

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func validate(v interface{}) []ValidationError {
	var errs []ValidationError
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := val.Field(i)
		tag := field.Tag.Get("validate")

		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			switch {
			case rule == "required":
				if value.IsZero() {
					errs = append(errs, ValidationError{
						Field:   field.Name,
						Message: "is required",
					})
				}
			case strings.HasPrefix(rule, "min="):
				var min int
				fmt.Sscanf(rule, "min=%d", &min)
				if value.Kind() == reflect.Int && value.Int() < int64(min) {
					errs = append(errs, ValidationError{
						Field:   field.Name,
						Message: fmt.Sprintf("must be >= %d", min),
					})
				}
			case strings.HasPrefix(rule, "max="):
				var max int
				fmt.Sscanf(rule, "max=%d", &max)
				if value.Kind() == reflect.Int && value.Int() > int64(max) {
					errs = append(errs, ValidationError{
						Field:   field.Name,
						Message: fmt.Sprintf("must be <= %d", max),
					})
				}
			}
		}
	}
	return errs
}

// =============================================================================
// PART 4: Dynamic function calls via reflection
// =============================================================================

func callMethod(obj interface{}, methodName string, args ...interface{}) []interface{} {
	val := reflect.ValueOf(obj)
	method := val.MethodByName(methodName)

	if !method.IsValid() {
		panic(fmt.Sprintf("method %s not found", methodName))
	}

	// Convert arguments to reflect.Value
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// Call the method
	results := method.Call(in)

	// Convert results back to interface{}
	out := make([]interface{}, len(results))
	for i, r := range results {
		out[i] = r.Interface()
	}
	return out
}

// =============================================================================
// PART 5: Creating types and values dynamically
// =============================================================================

func createDynamic() {
	fmt.Println("\n=== Dynamic Creation ===")

	// Create a new struct type at runtime
	fields := []reflect.StructField{
		{
			Name: "ID",
			Type: reflect.TypeOf(0),
			Tag:  `json:"id"`,
		},
		{
			Name: "Name",
			Type: reflect.TypeOf(""),
			Tag:  `json:"name"`,
		},
	}
	dynamicType := reflect.StructOf(fields)
	fmt.Printf("Dynamic type: %v\n", dynamicType)

	// Create an instance of the dynamic type
	instance := reflect.New(dynamicType).Elem()
	instance.Field(0).SetInt(42)
	instance.Field(1).SetString("Dynamic Object")
	fmt.Printf("Dynamic instance: %v\n", instance.Interface())

	// Create a slice of the dynamic type
	sliceType := reflect.SliceOf(dynamicType)
	slice := reflect.MakeSlice(sliceType, 0, 10)
	slice = reflect.Append(slice, instance)
	fmt.Printf("Dynamic slice: %v\n", slice.Interface())

	// Create a map dynamically
	mapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(0))
	m := reflect.MakeMap(mapType)
	m.SetMapIndex(reflect.ValueOf("key"), reflect.ValueOf(42))
	fmt.Printf("Dynamic map: %v\n", m.Interface())

	// Create a channel dynamically
	chanType := reflect.ChanOf(reflect.BothDir, reflect.TypeOf(0))
	ch := reflect.MakeChan(chanType, 5)
	ch.Send(reflect.ValueOf(100))
	v, _ := ch.Recv()
	fmt.Printf("Dynamic channel received: %v\n", v.Interface())
}

// =============================================================================
// PART 6: reflect.Select — Dynamic select statement
// =============================================================================

func dynamicSelect() {
	fmt.Println("\n=== Dynamic Select ===")

	ch1 := make(chan int, 1)
	ch2 := make(chan string, 1)
	ch2 <- "hello"

	cases := []reflect.SelectCase{
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch1)},
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch2)},
		{Dir: reflect.SelectDefault},
	}

	chosen, value, ok := reflect.Select(cases)
	fmt.Printf("Selected case %d: value=%v, ok=%v\n", chosen, value, ok)
}

func main() {
	user := User{Name: "Vikram", Email: "vikram@example.com", Age: 25}

	fmt.Println("=== Type Introspection ===")
	inspectType(user)

	inspectAndModify(&user)
	fmt.Printf("After modification: %+v\n", user)

	fmt.Println("\n=== Validation ===")
	badUser := User{Name: "", Age: -5}
	errs := validate(badUser)
	for _, err := range errs {
		fmt.Printf("  Validation error: %s\n", err)
	}

	fmt.Println("\n=== Dynamic Method Call ===")
	result := callMethod(&user, "Greet")
	fmt.Printf("Greet() = %v\n", result[0])

	callMethod(&user, "SetName", "New Name")
	fmt.Printf("After SetName: %s\n", user.Name)

	createDynamic()
	dynamicSelect()
}
