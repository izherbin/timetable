package pkg

func SliceContain(n int, sl []int) bool {
	for _, el := range sl {
		if el == n {
			return true
		}
	}
	return false
}
