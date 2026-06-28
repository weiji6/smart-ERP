package erp

func RoundPercent(amount Money, permille int64) Money {
	v := int64(amount) * permille
	if v >= 0 {
		return Money((v + 500) / 1000)
	}
	return Money((v - 500) / 1000)
}

func CeilPercent(amount Money, permille int64) Money {
	v := int64(amount) * permille
	if v <= 0 {
		return 0
	}
	return Money((v + 999) / 1000)
}

func FloorPercent(amount Money, permille int64) Money {
	v := int64(amount) * permille
	if v <= 0 {
		return 0
	}
	return Money(v / 1000)
}

func maxMoney(a, b Money) Money {
	if a > b {
		return a
	}
	return b
}

func minMoney(a, b Money) Money {
	if a < b {
		return a
	}
	return b
}
