package namec

var HoldWords = []string{"ID", "URL", "IP"}

func SnakeCase(name string) string {
	v := []byte(name)
	l := 0
	for i := 1; i < len(v); i++ {
		if 'A' <= v[i] && v[i] <= 'Z' {
			l++
		}
	}
	t := make([]byte, 0, l)
	for i := 0; i < len(v); i++ {
		if 'A' <= v[i] && v[i] <= 'Z' {
			applyHoldWords(v, i)
			if i > 0 {
				t = append(t, '_')
			}
			t = append(t, v[i]+('a'-'A'))
		} else {
			t = append(t, v[i])
		}
	}
	return string(t)
}

func applyHoldWords(name []byte, i int) {
	for _, hold := range HoldWords {
		if name[i] == hold[0] && i+len(hold) <= len(name) {
			for j := range hold {
				if name[i+j] != hold[j] {
					return
				}
			}
			for j := range hold {
				if j > 0 {
					name[i+j] = hold[j] + ('a' - 'A')
				}
			}
		}
	}
}

