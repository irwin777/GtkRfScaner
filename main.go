package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"github.com/tarm/serial"
	"log"
	"net/http"
	"time"
)

func main() {
	// Инициализируем GTK.
	gtk.Init(nil)

	// Создаём билдер
	b, err := gtk.BuilderNew()
	if err != nil {
		fmt.Println("Ошибка:", err)
	}

	// Загружаем в билдер окно из файла Glade
	err = b.AddFromFile("main.glade")
	if err != nil {
		fmt.Println("Ошибка:", err)
	}

	// Получаем объект главного окна по ID
	obj, err := b.GetObject("window_main")
	if err != nil {
		fmt.Println("Ошибка:", err)
	}

	// Преобразуем из объекта именно окно типа gtk.Window
	// и соединяем с сигналом "destroy" чтобы можно было закрыть
	// приложение при закрытии окна
	win := obj.(*gtk.Window)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Получаем поле ввода
	//obj, _ = b.GetObject("entry_1")
	//entry1 := obj.(*gtk.Entry)

	// Получаем кнопку
	obj, _ = b.GetObject("button_1")
	button1 := obj.(*gtk.Button)

	// Получаем метку
	obj, _ = b.GetObject("label_1")
	label1 := obj.(*gtk.Label)

	// Сигнал по нажатию на кнопку
	button1.Connect("clicked", func() {
		//rf_id, err := entry1.GetText()
		table := opros()
		if len(table) > 0 {
			var text string
			for k, v := range table {
				text += fmt.Sprintln(k, v)
			}
			label1.SetText(text)
		} else {
			label1.SetText("Метка не найдена")
		}
	})

	// Отображаем все виджеты в окне
	win.ShowAll()

	// Выполняем главный цикл GTK (для отрисовки). Он остановится когда
	// выполнится gtk.MainQuit()
	gtk.Main()
}

func crc(n int, buf []byte) byte {
	var uSum byte
	for i := 0; i < n-1; i++ {
		uSum += buf[i]
	}
	uSum = ^uSum + 1
	return uSum
}

func getnomer(rf_id string) string {
	url := "http://api.geo.astelitgroup.com/v1/rfid/"
	url = url + rf_id
	url = url + "/ref-dev-client"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return "error"
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Close = true
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "error"
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		ret := fmt.Sprintf("%v", result["refDevClnt"])
		return ret
	} else {
		return "none"
	}
}

func readport(port *serial.Port) (int, []byte) {
	buf0 := make([]byte, 512)
	n, err := port.Read(buf0)
	if err != nil {
		log.Fatal(err)
	}
	buf1 := make([]byte, n)
	buf1 = buf0[:n]
	crc := crc(n, buf1)
	if crc == buf1[n-1] {
		return n, buf1
	} else {
		return 0, []byte{0}
	}
}

func opros() map[string]string {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200, ReadTimeout: time.Second * 3}
	port, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()
	defer port.Flush()

	table := make(map[string]string)
	_, err = port.Write([]byte{0xA0, 0x04, 0x01, 0x89, 0x01, 0xD1})
	if err != nil {
		log.Fatal(err)
	}
	n, buf := readport(port)
	if n == 21 {
		for n == 21 {
			rf_id := hex.EncodeToString(buf[7:19])
			nomer := getnomer(rf_id)
			if _, ok := table[rf_id]; !ok {
				table[rf_id] = nomer
			}
			n, buf = readport(port)
		}
	}
	return table
}
