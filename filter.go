package builder

// Filter transforms an input string before it is converted and compared.
type Filter func(string) string
