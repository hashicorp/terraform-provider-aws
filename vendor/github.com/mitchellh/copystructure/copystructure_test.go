package copystructure

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestCopy_complex(t *testing.T) {
	v := map[string]interface{}{
		"foo": []string{"a", "b"},
		"bar": "baz",
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_interfacePointer(t *testing.T) {
	type Nested struct {
		Field string
	}

	type Test struct {
		Value *interface{}
	}

	ifacePtr := func(v interface{}) *interface{} {
		return &v
	}

	v := Test{
		Value: ifacePtr(Nested{Field: "111"}),
	}
	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_primitive(t *testing.T) {
	cases := []interface{}{
		42,
		"foo",
		1.2,
	}

	for _, tc := range cases {
		result, err := Copy(tc)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if result != tc {
			t.Fatalf("bad: %#v", result)
		}
	}
}

func TestCopy_primitivePtr(t *testing.T) {
	i := 42
	s := "foo"
	f := 1.2
	cases := []interface{}{
		&i,
		&s,
		&f,
	}

	for i, tc := range cases {
		result, err := Copy(tc)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(result, tc) {
			t.Fatalf("%d exptected: %#v\nbad: %#v", i, tc, result)
		}
	}
}

func TestCopy_map(t *testing.T) {
	v := map[string]interface{}{
		"bar": "baz",
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_array(t *testing.T) {
	v := [2]string{"bar", "baz"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_pointerToArray(t *testing.T) {
	v := &[2]string{"bar", "baz"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_slice(t *testing.T) {
	v := []string{"bar", "baz"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_pointerToSlice(t *testing.T) {
	v := &[]string{"bar", "baz"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_pointerToMap(t *testing.T) {
	v := &map[string]string{"bar": "baz"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_struct(t *testing.T) {
	type test struct {
		Value string
	}

	v := test{Value: "foo"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_structPtr(t *testing.T) {
	type test struct {
		Value string
	}

	v := &test{Value: "foo"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_structNil(t *testing.T) {
	type test struct {
		Value string
	}

	var v *test
	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if v, ok := result.(*test); !ok {
		t.Fatalf("bad: %#v", result)
	} else if v != nil {
		t.Fatalf("bad: %#v", v)
	}
}

func TestCopy_structNested(t *testing.T) {
	type TestInner struct{}

	type Test struct {
		Test *TestInner
	}

	v := Test{}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_structWithNestedArray(t *testing.T) {
	type TestInner struct {
		Value string
	}

	type Test struct {
		Value [2]TestInner
	}

	v := Test{
		Value: [2]TestInner{
			{Value: "bar"},
			{Value: "baz"},
		},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_structWithPointerToSliceField(t *testing.T) {
	type Test struct {
		Value *[]string
	}

	v := Test{
		Value: &[]string{"bar", "baz"},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_structWithPointerToArrayField(t *testing.T) {
	type Test struct {
		Value *[2]string
	}

	v := Test{
		Value: &[2]string{"bar", "baz"},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_structWithPointerToMapField(t *testing.T) {
	type Test struct {
		Value *map[string]string
	}

	v := Test{
		Value: &map[string]string{"bar": "baz"},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_structUnexported(t *testing.T) {
	type test struct {
		Value string

		private string
	}

	v := test{Value: "foo"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_structUnexportedMap(t *testing.T) {
	type Sub struct {
		Foo map[string]interface{}
	}

	type test struct {
		Value string

		private Sub
	}

	v := test{
		Value: "foo",
		private: Sub{
			Foo: map[string]interface{}{
				"yo": 42,
			},
		},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// private should not be copied
	v.private = Sub{}
	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad:\n\n%#v\n\n%#v", result, v)
	}
}

func TestCopy_structUnexportedArray(t *testing.T) {
	type Sub struct {
		Foo [2]string
	}

	type test struct {
		Value string

		private Sub
	}

	v := test{
		Value: "foo",
		private: Sub{
			Foo: [2]string{"bar", "baz"},
		},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// private should not be copied
	v.private = Sub{}
	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad:\n\n%#v\n\n%#v", result, v)
	}
}

// This is testing an unexported field containing a slice of pointers, which
// was a crashing case found in Terraform.
func TestCopy_structUnexportedPtrMap(t *testing.T) {
	type Foo interface{}

	type Sub struct {
		List []Foo
	}

	type test struct {
		Value string

		private *Sub
	}

	v := test{
		Value: "foo",
		private: &Sub{
			List: []Foo{&Sub{}},
		},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// private should not be copied
	v.private = nil
	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad:\n\n%#v\n\n%#v", result, v)
	}
}

func TestCopy_nestedStructUnexported(t *testing.T) {
	type subTest struct {
		mine string
	}

	type test struct {
		Value   string
		private subTest
	}

	v := test{Value: "foo"}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_time(t *testing.T) {
	type test struct {
		Value time.Time
	}

	v := test{Value: time.Now().UTC()}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestCopy_aliased(t *testing.T) {
	type (
		Int   int
		Str   string
		Map   map[Int]interface{}
		Slice []Str
	)

	v := Map{
		1: Map{10: 20},
		2: Map(nil),
		3: Slice{"a", "b"},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

type EmbeddedLocker struct {
	sync.Mutex
	Map map[int]int
}

func TestCopy_embeddedLocker(t *testing.T) {
	v := &EmbeddedLocker{
		Map: map[int]int{42: 111},
	}
	// start locked to prevent copying
	v.Lock()

	var result interface{}
	var err error

	copied := make(chan bool)

	go func() {
		result, err = Config{Lock: true}.Copy(v)
		close(copied)
	}()

	// pause slightly to make sure copying is blocked
	select {
	case <-copied:
		t.Fatal("copy completed while locked!")
	case <-time.After(100 * time.Millisecond):
		v.Unlock()
	}

	<-copied

	// test that the mutex is in the correct state
	result.(*EmbeddedLocker).Lock()
	result.(*EmbeddedLocker).Unlock()

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

// this will trigger the race detector, and usually panic if the original
// struct isn't properly locked during Copy
func TestCopy_lockRace(t *testing.T) {
	v := &EmbeddedLocker{
		Map: map[int]int{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				v.Lock()
				v.Map[i] = i
				v.Unlock()
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			Config{Lock: true}.Copy(v)
		}()
	}

	wg.Wait()
	result, err := Config{Lock: true}.Copy(v)

	// test that the mutex is in the correct state
	result.(*EmbeddedLocker).Lock()
	result.(*EmbeddedLocker).Unlock()

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

type LockedField struct {
	String string
	Locker *EmbeddedLocker
	// this should not get locked or have its state copied
	Mutex    sync.Mutex
	nilMutex *sync.Mutex
}

func TestCopy_lockedField(t *testing.T) {
	v := &LockedField{
		String: "orig",
		Locker: &EmbeddedLocker{
			Map: map[int]int{42: 111},
		},
	}

	// start locked to prevent copying
	v.Locker.Lock()
	v.Mutex.Lock()

	var result interface{}
	var err error

	copied := make(chan bool)

	go func() {
		result, err = Config{Lock: true}.Copy(v)
		close(copied)
	}()

	// pause slightly to make sure copying is blocked
	select {
	case <-copied:
		t.Fatal("copy completed while locked!")
	case <-time.After(100 * time.Millisecond):
		v.Locker.Unlock()
	}

	<-copied

	// test that the mutexes are in the correct state
	result.(*LockedField).Locker.Lock()
	result.(*LockedField).Locker.Unlock()
	result.(*LockedField).Mutex.Lock()
	result.(*LockedField).Mutex.Unlock()

	// this wasn't  blocking, but should be unlocked for DeepEqual
	v.Mutex.Unlock()

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("expected:\n%#v\nbad:\n%#v\n", v, result)
	}
}

// test something that doesn't contain a lock internally
type lockedMap map[int]int

var mapLock sync.Mutex

func (m lockedMap) Lock()   { mapLock.Lock() }
func (m lockedMap) Unlock() { mapLock.Unlock() }

func TestCopy_lockedMap(t *testing.T) {
	v := lockedMap{1: 2}
	v.Lock()

	var result interface{}
	var err error

	copied := make(chan bool)

	go func() {
		result, err = Config{Lock: true}.Copy(&v)
		close(copied)
	}()

	// pause slightly to make sure copying is blocked
	select {
	case <-copied:
		t.Fatal("copy completed while locked!")
	case <-time.After(100 * time.Millisecond):
		v.Unlock()
	}

	<-copied

	// test that the mutex is in the correct state
	result.(*lockedMap).Lock()
	result.(*lockedMap).Unlock()

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, &v) {
		t.Fatalf("bad: %#v", result)
	}
}

// Use an RLock if available
type RLocker struct {
	sync.RWMutex
	Map map[int]int
}

func TestCopy_rLocker(t *testing.T) {
	v := &RLocker{
		Map: map[int]int{1: 2},
	}
	v.Lock()

	var result interface{}
	var err error

	copied := make(chan bool)

	go func() {
		result, err = Config{Lock: true}.Copy(v)
		close(copied)
	}()

	// pause slightly to make sure copying is blocked
	select {
	case <-copied:
		t.Fatal("copy completed while locked!")
	case <-time.After(100 * time.Millisecond):
		v.Unlock()
	}

	<-copied

	// test that the mutex is in the correct state
	vCopy := result.(*RLocker)
	vCopy.Lock()
	vCopy.Unlock()
	vCopy.RLock()
	vCopy.RUnlock()

	// now make sure we can copy during an RLock
	v.RLock()
	result, err = Config{Lock: true}.Copy(v)
	if err != nil {
		t.Fatal(err)
	}
	v.RUnlock()

	vCopy = result.(*RLocker)
	vCopy.Lock()
	vCopy.Unlock()
	vCopy.RLock()
	vCopy.RUnlock()

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("bad: %#v", result)
	}
}

// Test that we don't panic when encountering nil Lockers
func TestCopy_missingLockedField(t *testing.T) {
	v := &LockedField{
		String: "orig",
	}

	result, err := Config{Lock: true}.Copy(v)

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("expected:\n%#v\nbad:\n%#v\n", v, result)
	}
}

type PointerLocker struct {
	Mu sync.Mutex
}

func (p *PointerLocker) Lock()   { p.Mu.Lock() }
func (p *PointerLocker) Unlock() { p.Mu.Unlock() }

func TestCopy_pointerLockerNil(t *testing.T) {
	v := struct {
		P *PointerLocker
	}{}

	_, err := Config{Lock: true}.Copy(&v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestCopy_sliceWithNil(t *testing.T) {
	v := [](*int){nil}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", v, result)
	}
}

func TestCopy_mapWithNil(t *testing.T) {
	v := map[int](*int){0: nil}

	result, err := Copy(v)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(result, v) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", v, result)
	}
}

// While this is safe to lock and copy directly, copystructure requires a
// pointer to reflect the value safely.
func TestCopy_valueWithLockPointer(t *testing.T) {
	v := struct {
		*sync.Mutex
		X int
	}{
		Mutex: &sync.Mutex{},
		X:     3,
	}

	_, err := Config{Lock: true}.Copy(v)

	if err != errPointerRequired {
		t.Fatalf("expected errPointerRequired, got: %v", err)
	}
}

func TestCopy_mapWithPointers(t *testing.T) {
	type T struct {
		S string
	}
	v := map[string]interface{}{
		"a": &T{S: "hello"},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatalf("%#v", result)
	}
}

func TestCopy_structWithMapWithPointers(t *testing.T) {
	type T struct {
		S string
		M map[string]interface{}
	}
	v := &T{
		S: "a",
		M: map[string]interface{}{
			"b": &T{
				S: "b",
			},
		},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatal(result)
	}
}

type testT struct {
	N   int
	Spp **string
	X   testX
	Xp  *testX
	Xpp **testX
}

type testX struct {
	Tp  *testT
	Tpp **testT
	Ip  *interface{}
	Ep  *error
	S   fmt.Stringer
}

type stringer struct{}

func (s *stringer) String() string {
	return "test string"
}

func TestCopy_structWithPointersAndInterfaces(t *testing.T) {
	// test that we can copy various nested and chained pointers and interfaces
	s := "val"
	sp := &s
	spp := &sp
	i := interface{}(11)

	tp := &testT{
		N: 2,
	}

	xp := &testX{
		Tp:  tp,
		Tpp: &tp,
		Ip:  &i,
		S:   &stringer{},
	}

	v := &testT{
		N:   1,
		Spp: spp,
		X:   testX{},
		Xp:  xp,
		Xpp: &xp,
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatal(result)
	}
}

func Test_pointerInterfacePointer(t *testing.T) {
	s := "hi"
	si := interface{}(&s)
	sip := &si

	result, err := Copy(sip)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(sip, result) {
		t.Fatalf("%#v != %#v\n", sip, result)
	}
}

func Test_pointerInterfacePointer2(t *testing.T) {
	type T struct {
		I *interface{}
		J **fmt.Stringer
	}

	x := 1
	y := &stringer{}

	i := interface{}(&x)
	j := fmt.Stringer(y)
	jp := &j

	v := &T{
		I: &i,
		J: &jp,
	}
	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatalf("%#v != %#v\n", v, result)
	}
}

// This test catches a bug that happened when unexported fields were
// first their subsequent fields wouldn't be copied.
func TestCopy_unexportedFieldFirst(t *testing.T) {
	type P struct {
		mu       sync.Mutex
		Old, New string
	}

	type T struct {
		M map[string]*P
	}

	v := &T{
		M: map[string]*P{
			"a": &P{Old: "", New: "2"},
		},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatalf("\n%#v\n\n%#v", v, result)
	}
}

func TestCopy_nilPointerInSlice(t *testing.T) {
	type T struct {
		Ps []*int
	}

	v := &T{
		Ps: []*int{nil},
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatalf("\n%#v\n\n%#v", v, result)
	}
}

//-------------------------------------------------------------------
// The tests below all tests various pointer cases around copying
// a structure that uses a defined Copier. This was originally raised
// around issue #26.

func TestCopy_timePointer(t *testing.T) {
	type T struct {
		Value *time.Time
	}

	now := time.Now()
	v := &T{
		Value: &now,
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatalf("\n%#v\n\n%#v", v, result)
	}
}

func TestCopy_timeNonPointer(t *testing.T) {
	type T struct {
		Value time.Time
	}

	v := &T{
		Value: time.Now(),
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatalf("\n%#v\n\n%#v", v, result)
	}
}

func TestCopy_timeDoublePointer(t *testing.T) {
	type T struct {
		Value **time.Time
	}

	now := time.Now()
	nowP := &now
	nowPP := &nowP
	v := &T{
		Value: nowPP,
	}

	result, err := Copy(v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, result) {
		t.Fatalf("\n%#v\n\n%#v", v, result)
	}
}
