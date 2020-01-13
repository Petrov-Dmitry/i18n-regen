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
    localesEntryPoint       = "https://api-sol.kube.dev001.ru/api/cms/locales"
    translationsEntryPoint  = "https://api-sol.kube.dev001.ru/api/cms/strings/%locale%"
)

type localesList struct {
    code, name, iso string
    isDefault bool
    nameInLocale string
}

func main() {
    fmt.Printf("\n=============================\nStart translations rebuild...\n=============================\n\n")

    // Если список локалей еще не загружен в файл
    if !fileExists(localesFile) {
        localesList := loadLocales()
        log.Println("> locales list loaded and saved")
        fmt.Println(localesList)
    } else {
        log.Println("> locales list already saved")
    }

    // Для каждой из полученных локалей, кроме русской
        // Получить переводы строк для локали из https://api-sol.kube.dev001.ru/api/cms/strings/%locale%
        // Создать файл %locale%.v2.json
        // Читаем блоки и вложенные в них строки-v2 (неограниченная вложенность) из файла ru.v2.json
        // Для каждой найденной строки
            // Находим соответствующий строке-v2 ключ-v1 в файле ru.v1.json
            // По найденному ключу-v1 находим значение перевода строки для локали
            // Записываем ключ-v2 и перевод строки в файл %locale%.v2.json
            // Проверяем наличие в логе создания файла записи локаль + ключ-v1 - если находим, то пишем в SQLite лог повторяющихся ключей (их надо будет проверить вручную)
            // Записываем в SQLite лог создания файла перевода - локаль + ключ-v1 + ключ-v2 + перевод

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
func loadLocales() string {
    // Получить список локалей из https://api-sol.kube.dev001.ru/api/cms/locales
    log.Println("> Try to load locales list...")
    localesResp, err := http.Get(localesEntryPoint)
    if err != nil {
        log.Fatalln(err)
    }
    if localesResp.StatusCode != 200 {
        log.Fatalln("\t>> 404 - Not found")
    }
    log.Println("\t>> statusCode is " + strconv.Itoa(localesResp.StatusCode))

    defer localesResp.Body.Close()
    body, err := ioutil.ReadAll(localesResp.Body)
    if err != nil {
        log.Fatalln(err)
    }
    log.Println("\t>> body is ")
    fmt.Printf("%+v", string([]byte(body)) + "\n")

    return string(body)
}
