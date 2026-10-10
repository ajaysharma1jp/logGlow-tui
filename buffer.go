package main

type LogBuffer struct {
	entries []LogEntry
	head int
	tail int
	count int
	cap int
}

func NewLogBuffer(capacity int) *LogBuffer{
	if capacity<=0 {
		panic("log buffer capacity must be positive")
	}
	return &LogBuffer{
		entries: make([]LogEntry, capacity),
		cap: capacity,
	}
}

func (buff *LogBuffer) Add(log LogEntry){
	buff.entries[buff.tail] = log
	buff.tail = (buff.tail+1)%buff.cap
	if buff.count < buff.cap{
		buff.count++
	}else{
		buff.head = (buff.head+1)%buff.cap
	}
}

func (buff *LogBuffer) Len() int{
	return buff.count
}

// at returns the log at a logical index (0 is oldest, Len()-1 is newest)
func (buff *LogBuffer) At(logicalIndex int) LogEntry{
	physicalIndex := (buff.head+logicalIndex)%buff.cap
	return buff.entries[physicalIndex]
}