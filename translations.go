// Генерация файлов переводов после смены структуры ключей строк i18n
package main

import (
    "fmt"
    "log"
    "os"
    "io/ioutil"
    "net/http"
    "strconv"
    "reflect"
    "encoding/json"
)

const (
    localesFile             = "locales.json"
    localesEntryPoint       = "https://api-sol.kube.dev001.ru/api/cms/locales/dev"
    translationsEntryPoint  = "https://api-sol.kube.dev001.ru/api/cms/strings/"
)

var (
    iteratedKeyParents []int
    iteratedKeyPath string
)

type LocalesList struct {
    LocalesList []Locale `json:"data"`
    DataUpdatedAt string `json:"dataUpdatedAt"`
}

type Locale struct {
    Code            string `json:"code"`
    Name            string `json:"name"`
    Iso             string `json:"iso"`
    IsDefault       bool   `json:"default"`
    NameInLocale    string `json:"name_in_locale"`
}

func main() {
    fmt.Printf("\n=============================\nStart translations rebuild...\n=============================\n\n")

    // Если список локалей еще не загружен в файл
    if !fileExists(localesFile) {
        if loadLocalesFile() {
            log.Println("> locales file loaded")
        } else {
            log.Fatalln("> locales file is NOT loaded")
        }
    } else {
        log.Println("> locales file already saved")
    }

    // Читаем список локалей из файла
    log.Println("> parse locales file...")
    file, _ := ioutil.ReadFile(localesFile)
    locales := LocalesList{}
    _ = json.Unmarshal([]byte(file), &locales)
    log.Println(" >> file parsed,", len(locales.LocalesList), "locales found")
    // fmt.Println(locales.LocalesList)

    // Для каждой из полученных локалей, кроме русской
    for i := 0; i < len(locales.LocalesList); i++ {
        log.Println("> process locale", locales.LocalesList[i].Code, locales.LocalesList[i].Name, locales.LocalesList[i].NameInLocale)

        // Получить переводы строк для локали из https://api-sol.kube.dev001.ru/api/cms/strings/%locale%
        if !loadLocaleFile(locales.LocalesList[i].Code) {
            log.Fatalln(" >> locale file is NOT loaded")
        }

        // Пересоздать файл locale.%locale%.v2.json
        locale := locales.LocalesList[i].Code
        localeV2File := "locale." + locale + ".v2.json"
        if fileExists(localeV2File) {
            err := os.Remove(localeV2File)
            if err != nil { log.Fatalln(err) }
            log.Println(" >> old v2 file deleted")
        }
        log.Println(" >> create new v2 file", locale)
        file, err := os.Create(localeV2File)
        if err != nil { log.Fatalln(err) }
        defer file.Close()
        log.Println(" >> new v2 file created", locale)

        // Читаем из файла ru.v2.json блоки и вложенные в них ключи-v2 (неограниченная вложенность)
        // v2file, err := os.Open("ru.v1.json")
        v2file, err := os.Open("ru.v2.json")
        if err != nil { log.Fatalln(err) }
        defer v2file.Close()
        log.Println(" >> source v2 file opened")
        v2ByteValue, _ := ioutil.ReadAll(v2file)
        var sourceV2 map[string]interface{}
        json.Unmarshal([]byte(v2ByteValue), &sourceV2)
        log.Println(" >> source v2 file parsed")
        // fmt.Println(sourceV2)

        // Для каждой найденной строки
        iterate(sourceV2)
        log.Println(" >> source v2 file iterated")
            // Находим соответствующий строке-v2 ключ-v1 в файле ru.v1.json
            // По найденному ключу-v1 находим значение перевода строки для локали
            // Записываем ключ-v2 и перевод строки в файл %locale%.v2.json
            // Проверяем наличие в логе создания файла записи локаль + ключ-v1 - если находим, то пишем в SQLite лог повторяющихся ключей (их надо будет проверить вручную)
            // Записываем в SQLite лог создания файла перевода - локаль + ключ-v1 + ключ-v2 + перевод
    }
    fmt.Printf("\n=========================\nTranslations rebuild done\n=========================\n")
}

// Рекурсивный спуск по ключам JSON
func iterate(data interface{}) interface{} {
    if reflect.ValueOf(data).Kind() == reflect.Slice {
        log.Println("  >>> slise found")
        os.Exit(1)
        var returnSlice []interface{}
        d := reflect.ValueOf(data)
        tmpData := make([]interface{}, d.Len())
        for i := 0; i < d.Len(); i++ {
            tmpData[i] = d.Index(i).Interface()
        }
        for i, v := range tmpData {
            returnSlice[i] = iterate(v)
        }
        return returnSlice
    } else if reflect.ValueOf(data).Kind() == reflect.Map {
        log.Println("  >>> map found")
        d := reflect.ValueOf(data)
        tmpData := make(map[string]interface{})
        for _, k := range d.MapKeys() {
            // Собираем цепочку родительских ключей
            if iteratedKeyPath != "" {
                iteratedKeyPath = iteratedKeyPath + "." + k.String()
            } else {
                iteratedKeyPath = k.String()
            }
            log.Println("  >>> prarse key", k, "in", iteratedKeyPath)
            typeOfValue := reflect.TypeOf(d.MapIndex(k).Interface()).Kind()
            if typeOfValue == reflect.Map || typeOfValue == reflect.Slice {
                // Поймали слайс - надо спускаться ниже
                log.Println("   >>>> that is slise", d.MapIndex(k))
                tmpData[k.String()] = iterate(d.MapIndex(k).Interface())
            } else {
                // Поймали строку перевода
                log.Println("   >>>> that is", typeOfValue, d.MapIndex(k))
                tmpData[k.String()] = d.MapIndex(k).Interface()
                // Убираем последний ключ из цепочки родительских ключей
                iteratedKeyPath = ""
            }
        }
        return tmpData
    }
    return data
}

// Загрузка файла локализации из API
func loadLocaleFile(locale string) bool {
    localeFile := "locale." + locale + ".json"
    // Получить переводы строк для локали из https://api-sol.kube.dev001.ru/api/cms/strings/%locale%
    log.Println(" >> Try to load locale file...", locale, translationsEntryPoint + locale, localeFile)
    if fileExists(localeFile) {
        log.Println(" >> locale file already loaded")
        return true
    }
    localeResp, err := http.Get(translationsEntryPoint + locale)
    if err != nil { log.Fatalln(err) }
    if localeResp.StatusCode != 200 {
        log.Fatalln("  >>> 404 - Not found")
    }
    log.Println("  >>> statusCode is", strconv.Itoa(localeResp.StatusCode))

    // Читаем полученный ответ API
    defer localeResp.Body.Close()
    body, err := ioutil.ReadAll(localeResp.Body)
    if err != nil { log.Fatalln(err) }
    log.Println("  >>> body is")
    // fmt.Println(string([]byte(body)))

    // Пишем ответ в файл
    f, err := os.Create(localeFile)
    if err != nil { log.Fatalln(err) }
    l, err := f.WriteString(string([]byte(body)))
    if err != nil { log.Fatalln(err) }
    log.Println("  >>> bytes written", l)
    err = f.Close()
    if err != nil { log.Fatalln(err) }
    log.Println("  >>> file closed", localeFile)

    return true
}

// loadLocales collect localesList from API and store it into the file locales.json
func loadLocalesFile() bool {
    // Получить список локалей из https://api-sol.kube.dev001.ru/api/cms/locales
    log.Println("> Try to load locales file...")
    localesResp, err := http.Get(localesEntryPoint)
    if err != nil { log.Fatalln(err) }
    if localesResp.StatusCode != 200 {
        log.Fatalln(" >> 404 - Not found")
    }
    log.Println(" >> statusCode is", strconv.Itoa(localesResp.StatusCode))

    // Читаем полученный ответ API
    defer localesResp.Body.Close()
    body, err := ioutil.ReadAll(localesResp.Body)
    if err != nil { log.Fatalln(err) }
    log.Println(" >> body is")
    // fmt.Println(string([]byte(body)))

    // Пишем ответ в файл
    f, err := os.Create(localesFile)
    if err != nil { log.Fatalln(err) }
    l, err := f.WriteString(string([]byte(body)))
    if err != nil { log.Fatalln(err) }
    log.Println(" >> bytes written", l)
    err = f.Close()
    if err != nil { log.Fatalln(err) }
    log.Println(" >> file closed")

    return true
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
