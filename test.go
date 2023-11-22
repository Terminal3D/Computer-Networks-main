package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Открытие файла для записи
	file, err := os.OpenFile("output.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	// Создание буферизованного writer'a для записи в файл
	writer := bufio.NewWriter(file)

	input := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Введите строку (или 'exit' для выхода):") // Подсказка для пользователя
		input.Scan()
		text := input.Text()

		// Проверка на команду выхода
		if text == "exit" {
			break
		}

		// Предполагается, что входные данные соответствуют ожидаемому формату
		if len(strings.Split(text, "\"")) > 1 {
			creatureName := strings.Split(text, "\"")[1]
			modifiedName := strings.Replace(creatureName, ":", "_", -1)
			modifiedName = "- " + modifiedName
			// Запись в файл вместо вывода на экран
			_, err := writer.WriteString(modifiedName + "\n")
			if err != nil {
				fmt.Println("Ошибка при записи в файл:", err)
				return
			}
			writer.Flush() // Очистка буфера для гарантии записи в файл
		} else {
			fmt.Println("Введенная строка не соответствует ожидаемому формату.")
		}
	}
}
