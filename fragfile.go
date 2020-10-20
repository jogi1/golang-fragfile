package fragfile

import "fmt"
import "os"
import "io/ioutil"
import "bytes"
import "bufio"
import "strings"
import "regexp"
import "reflect"

//import "github.com/jogi1/mvdreader"

func (fragfile *Fragfile) Define(token string, values []string) error {
	var define interface{}
	switch token {
	case "WEAPON_CLASS":
		fallthrough
	case "WC":
		wc := new(WeaponClass)
		define = wc
		fragfile.WeaponClass[values[0]] = append(fragfile.WeaponClass[values[0]], wc)
		break
	case "OBITUARY":
		fallthrough
	case "OBIT":
		obit := new(Obituary)
		define = obit
		fragfile.Obituary[values[0]] = append(fragfile.Obituary[values[0]], obit)
		break
	case "FLAG_ALERT":
		fallthrough
	case "FLAG_MSG":
		flagAlert := new(FlagAlert)
		define = flagAlert
		fragfile.FlagAlert[values[0]] = append(fragfile.FlagAlert[values[0]], flagAlert)
		break
	default:
		return fmt.Errorf("token \"%s\" not understood.", token)
	}
	elem := reflect.ValueOf(define).Elem()
	elemType := reflect.TypeOf(define).Elem()
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		etf := elemType.Field(i)
		required := etf.Tag == "required"
		if i >= len(values) {
			if required {
				return fmt.Errorf("required field \"%s\" not set %s, %s", etf.Name, token, values)
			}
			break
		}
		val := values[i]
		if etf.Name == "Message1" || etf.Name == "Message2" {
			val = regexp.QuoteMeta(val)
		}
		if strings.HasPrefix(val, "//") {
			if required {
				return fmt.Errorf("required field \"%s\" can not be set with comment", etf.Name)
			}
			break
		}
		if f.CanSet() {
			f.SetString(val)
		}
	}
	return nil
}

type WeaponClass struct {
	Keyword   string `required`
	Name      string `required`
	ShortName string
}

type Obituary struct {
	Type     string `required`
	Weapon   string `required`
	Message1 string `required`
	Message2 string
}

type FlagAlert struct {
	Type     string `required`
	Message1 string `required`
}

type Fragfile struct {
	Info        map[string]string
	Meta        map[string]string
	WeaponClass map[string][]*WeaponClass
	Obituary    map[string][]*Obituary
	FlagAlert   map[string][]*FlagAlert
}

type FragMessage struct {
	X, Y, Type, Weapon string
}

func (fragfile *Fragfile) ParseMessage(message string) (*FragMessage, error) {
	fm := new(FragMessage)
	for name, types := range fragfile.Obituary {
		var s string
		for _, typ := range types {
			if len(typ.Message2) > 0 {
				s = fmt.Sprintf("(.*)%s(.*)%s", typ.Message1, typ.Message2)
			} else {
				s = fmt.Sprintf("(.*)%s(.*)", typ.Message1)
			}
			r, err := regexp.Compile(s)
			if err != nil {
				fmt.Println(s)
				return nil, err
			}
			match := r.FindStringSubmatch(message)
			//if r.MatchString(message) {
			if len(match) > 0 {
				if len(match) > 1 {
					fm.X = match[1]
				}
				if len(match) > 2 {
					fm.Y = match[2]
				}
				fm.Type = name
				fm.Weapon = typ.Weapon
				return fm, nil
			}
		}
	}
	return nil, nil
}

func FragfileLoadByte(data []byte) (*Fragfile, error) {
	fragfile := new(Fragfile)
	fragfile.Info = make(map[string]string)
	fragfile.Meta = make(map[string]string)
	fragfile.Meta = make(map[string]string)
	fragfile.WeaponClass = make(map[string][]*WeaponClass)
	fragfile.Obituary = make(map[string][]*Obituary)
	fragfile.FlagAlert = make(map[string][]*FlagAlert)
	bytesReader := bytes.NewReader(data)
	bufReader := bufio.NewReader(bytesReader)

	for line, _, _ := bufReader.ReadLine(); line != nil; line, _, _ = bufReader.ReadLine() {
		s := string(line)
		if strings.HasPrefix(s, "//") {
			continue
		}
		r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
		splits := r.FindAllString(s, -1)
		if len(splits) < 1 {
			continue
		}
		for x, split := range splits {
			split = strings.Trim(split, "\"")
			splits[x] = split
		}
		token := splits[0]
		switch token {
		case "#FRAGFILE":
			fragfile.Info[splits[1]] = splits[2]
			continue

		case "#META":
			fragfile.Meta[splits[1]] = splits[2]
			continue
		case "#DEFINE":
			defineToken := splits[1]
			err := fragfile.Define(defineToken, splits[2:])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			continue
		default:
			fmt.Println("unknown token: ", token)
			os.Exit(1)
			break
		}
	}
	return fragfile, nil
}

func FragfileLoadFile(filename string) (*Fragfile, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return FragfileLoadByte(data)
}

/*
func main() {
	fragfile, err := FragfileLoadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	read_file, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err, mvd := mvdreader.Load(read_file, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for err, done := mvd.ParseFrame(); done != true && err == nil; err, done = mvd.ParseFrame() {
		for _, message := range mvd.State.Messages {
			fm, err := fragfile.ParseMessage(message.Message)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if fm != nil {
				fmt.Println(fm)
			}
		}
	}
	//fmt.Println(fragfile)
}
*/
