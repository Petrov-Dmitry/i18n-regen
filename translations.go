// Генерация файлов переводов после смены структуры ключей строк i18n
package main

import (
    "fmt"
    "log"
    "os"
    "io/ioutil"
    "net/http"
    "strconv"
    "encoding/json"
)

const (
    localesFile             = "locales.json"
    localesEntryPoint       = "https://api-sol.kube.dev001.ru/api/cms/locales/dev"
    translationsEntryPoint  = "https://api-sol.kube.dev001.ru/api/cms/strings/"
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
    fmt.Println(locales.LocalesList)

    // Для каждой из полученных локалей, кроме русской
    for i := 0; i < len(locales.LocalesList); i++ {
        log.Println("> process locale", locales.LocalesList[i].Code, locales.LocalesList[i].Name, locales.LocalesList[i].NameInLocale)

        // Получить переводы строк для локали из https://api-sol.kube.dev001.ru/api/cms/strings/%locale%
        if !loadLocaleFile(locales.LocalesList[i].Code) {
            log.Fatalln(" >> locale file is NOT loaded")
        }

        // Создать файл %locale%.v2.json
        // Читаем блоки и вложенные в них строки-v2 (неограниченная вложенность) из файла ru.v2.json
        // Для каждой найденной строки
            // Находим соответствующий строке-v2 ключ-v1 в файле ru.v1.json
            // По найденному ключу-v1 находим значение перевода строки для локали
            // Записываем ключ-v2 и перевод строки в файл %locale%.v2.json
            // Проверяем наличие в логе создания файла записи локаль + ключ-v1 - если находим, то пишем в SQLite лог повторяющихся ключей (их надо будет проверить вручную)
            // Записываем в SQLite лог создания файла перевода - локаль + ключ-v1 + ключ-v2 + перевод
    }
    fmt.Printf("\n=========================\nTranslations rebuild done\n=========================\n")
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

// loadLocales collect localesList from API and store it into the file locales.json
func loadLocalesFile() bool {
    // Получить список локалей из https://api-sol.kube.dev001.ru/api/cms/locales
    log.Println("> Try to load locales file...")
    localesResp, err := http.Get(localesEntryPoint)
    if err != nil {
        log.Fatalln(err)
    }
    if localesResp.StatusCode != 200 {
        log.Fatalln(" >> 404 - Not found")
    }
    log.Println(" >> statusCode is", strconv.Itoa(localesResp.StatusCode))

    // Читаем полученный ответ API
    defer localesResp.Body.Close()
    body, err := ioutil.ReadAll(localesResp.Body)
    if err != nil {
        log.Fatalln(err)
    }
    log.Println(" >> body is")
    fmt.Println(string([]byte(body)))

    // Пишем ответ в файл
    f, err := os.Create(localesFile)
    if err != nil {
        log.Fatalln(err)
    }
    l, err := f.WriteString(string([]byte(body)))
    if err != nil {
        log.Fatalln(err)
    }
    log.Println(" >> bytes written", l)
    err = f.Close()
    if err != nil {
        log.Fatalln(err)
    }
    log.Println(" >> file closed")

    return true
}

func loadLocaleFile(locale string) bool {
    var localeFile = "locale." + locale + ".json"
    // Получить переводы строк для локали из https://api-sol.kube.dev001.ru/api/cms/strings/%locale%
    log.Println(" >> Try to load locale file...", locale, translationsEntryPoint + locale, localeFile)
    if fileExists(localeFile) {
        log.Println(" >> locale file already loaded")
        return true
    }
    localeResp, err := http.Get(translationsEntryPoint + locale)
    if err != nil {
        log.Fatalln(err)
    }
    if localeResp.StatusCode != 200 {
        log.Fatalln("  >>> 404 - Not found")
    }
    log.Println("  >>> statusCode is", strconv.Itoa(localeResp.StatusCode))

    // Читаем полученный ответ API
    defer localeResp.Body.Close()
    body, err := ioutil.ReadAll(localeResp.Body)
    if err != nil {
        log.Fatalln(err)
    }
    log.Println("  >>> body is")
    fmt.Println(string([]byte(body)))

    // Пишем ответ в файл
    f, err := os.Create(localeFile)
    if err != nil {
        log.Fatalln(err)
    }
    l, err := f.WriteString(string([]byte(body)))
    if err != nil {
        log.Fatalln(err)
    }
    log.Println("  >>> bytes written", l)
    err = f.Close()
    if err != nil {
        log.Fatalln(err)
    }
    log.Println("  >>> file closed", localeFile)

    return true
}
