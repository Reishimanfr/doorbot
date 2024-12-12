package utils

func RepeatString(str string, amount int) string {
	output := ""
	for i := 0; i < amount; i++ {
		output += str
	}
	return output
}
