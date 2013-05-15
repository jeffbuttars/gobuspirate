
package buspirate

import "testing"

func TestNewBP (t *testing.T) {

    nbp := NewBP("")

    if nil == nbp {
        t.Fatalf("Unable to allocate an Buspirate IO instance")
    }

    // Make sure the default values are what we expect
    if nbp.Device != "/dev/buspirate" {
        t.Fatalf("Invalid default buspirate Device: %s", nbp.Device)
    }

    dev := "/dev/ttyUSB0"
    nbp = NewBP(dev)
    if nbp.Device != dev {
        t.Fatalf("Invalid buspirate Device: %s, expected: ", nbp.Device, dev)
    }
} //TestNewBP()

func TestBPInit (t *testing.T) {

    nbp := NewBP("")
    if nil == nbp {
        t.Fatalf("Unable to allocate an Buspirate IO instance")
    }

    // Make sure we have a Serial object
    err := nbp.Init()
    if err != nil {
        t.Fail()
        t.Logf("Unable to initialize a Buspirate IO instance\n%s", err)
        return
    }

    if nbp.Serial == nil {
        t.Fatalf("Unable to initialize a Buspirate Serial instance")
    }

    // make sure the buffer is the correct size
    if len(nbp.buf) != BUF_SIZE {
        t.Fatalf("Expected a buf size of %d, got %d", BUF_SIZE, len(nbp.buf))
    }
} //TestBPInit()
