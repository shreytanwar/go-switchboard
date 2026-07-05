type Store struct {
	mu sync.RWMutex
	data map[string][]byte // topic to latest value
}

func (s *Store) Set(topic string, value []byte)
func (s *Store) Get(topic string) ([]byte, bool)

//Short poll reads from here. Publish endpoint writes to both the broker (streaming transports) and the store (short poll).