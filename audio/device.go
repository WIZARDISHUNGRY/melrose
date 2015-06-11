package audio

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	m "github.com/emicklei/melrose"
	"github.com/vova616/go-openal/openal"
)

type Device struct {
	Name         string
	soundContext *openal.Context
	waves        map[string]openal.Buffer
}

func (d *Device) Open() {
	device := openal.OpenDevice(d.Name)
	context := device.CreateContext()
	context.Activate()
	d.soundContext = context
	d.waves = map[string]openal.Buffer{}
}

func (d *Device) Close() {
	if d.soundContext != nil {
		d.soundContext.Destroy()
	}
}

// half a second
func (d *Device) PlayNote(note m.Note, duration time.Duration) {
	key := note.Whole().String()
	wav, ok := d.waves[key]
	if !ok {
		log.Printf("No such note:%s", key)
		return
	}
	//fmt.Println("=", key)
	source := openal.NewSource()
	source.SetPitch(1)
	source.SetGain(1)
	source.SetPosition(0, 0, 0)
	source.SetVelocity(0, 0, 0)
	source.SetLooping(false)
	source.SetBuffer(wav)
	source.Play()
	time.Sleep(duration)
	source.Stop()
}

func (d *Device) LoadSounds() {
	dir := filepath.Join(os.Getenv("HOME"), "sounds")
	list, _ := ioutil.ReadDir(dir)
	for _, each := range list {
		if strings.HasSuffix(each.Name(), ".wav") {
			fin := path.Join(dir, each.Name())
			d.loadWavFile(fin)
		}
	}
	log.Printf("loaded %d sound files\n", len(d.waves))
}

func (d *Device) loadWavFile(fileName string) {
	buffer := openal.NewBuffer()
	format, data, err := readWavFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	switch format.Channels {
	case 1:
		buffer.SetData(openal.FormatMono16, data[:len(data)], int32(format.Samples))
	case 2:
		buffer.SetData(openal.FormatStereo16, data[:len(data)], int32(format.Samples))
	}
	key := path.Base(fileName)                           // gs3.wav
	key = strings.ToUpper(key[0 : len(key)-len(".wav")]) // GS3
	key = strings.Replace(key, "S", "♯", -1)
	note := m.ParseNote(key)
	if note.IsSharp() {
		flat := note.Pitched(1).Modified(m.Flat)
		//fmt.Println(flat, "->", fileName)
		d.waves[flat.String()] = buffer
	}
	//fmt.Println(note, "->", fileName)
	d.waves[note.String()] = buffer
}

type Format struct {
	FormatTag     int16
	Channels      int16
	Samples       int32
	AvgBytes      int32
	BlockAlign    int16
	BitsPerSample int16
}

type Format2 struct {
	Format
	SizeOfExtension int16
}

type Format3 struct {
	Format2
	ValidBitsPerSample int16
	ChannelMask        int32
	SubFormat          [16]byte
}

func readWavFile(path string) (*Format, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var buff [4]byte
	f.Read(buff[:4])

	if string(buff[:4]) != "RIFF" {
		return nil, nil, fmt.Errorf("Not a WAV file.\n")
	}

	var size int32
	binary.Read(f, binary.LittleEndian, &size)

	f.Read(buff[:4])

	if string(buff[:4]) != "WAVE" {
		return nil, nil, fmt.Errorf("Not a WAV file.\n")
	}

	f.Read(buff[:4])

	if string(buff[:4]) != "fmt " {
		return nil, nil, fmt.Errorf("Not a WAV file.\n")
	}

	binary.Read(f, binary.LittleEndian, &size)

	var format Format

	switch size {
	case 16:
		binary.Read(f, binary.LittleEndian, &format)
	case 18:
		var f2 Format2
		binary.Read(f, binary.LittleEndian, &f2)
		format = f2.Format
	case 40:
		var f3 Format3
		binary.Read(f, binary.LittleEndian, &f3)
		format = f3.Format
	}

	//fmt.Println(format)

	f.Read(buff[:4])

	if string(buff[:4]) != "data" {
		return nil, nil, fmt.Errorf("Not supported WAV file.\n")
	}

	binary.Read(f, binary.LittleEndian, &size)

	wavData := make([]byte, size)
	n, e := f.Read(wavData)
	if e != nil {
		return nil, nil, fmt.Errorf("Cannot read WAV data.\n")
	}
	if int32(n) != size {
		return nil, nil, fmt.Errorf("WAV data size doesnt match.\n")
	}

	return &format, wavData, nil
}
