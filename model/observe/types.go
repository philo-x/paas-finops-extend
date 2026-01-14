package observe

import (
	"database/sql/driver"
	"time"
)

// NullTime 可空时间类型
type NullTime struct {
	*time.Time
}

// Value 实现 driver.Valuer 接口，写入数据库前调用
func (nt NullTime) Value() (driver.Value, error) {
	if nt.Time == nil || nt.Time.IsZero() {
		return nil, nil // 如果是 0001-01-01，存入数据库为 NULL
	}
	return *nt.Time, nil
}

// Scan 实现 sql.Scanner 接口，从数据库读取时调用
func (nt *NullTime) Scan(v interface{}) error {
	t, ok := v.(time.Time)
	if ok {
		nt.Time = &t
	}
	return nil
}
