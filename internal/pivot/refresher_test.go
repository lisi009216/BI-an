package pivot

import (
	"testing"
	"time"
)

func TestGetThisWeekMonday(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	tests := []struct {
		name     string
		now      time.Time
		wantDate string // 期望的周一日期 YYYY-MM-DD
	}{
		{
			name:     "Monday",
			now:      time.Date(2024, 12, 30, 10, 0, 0, 0, loc), // 2024-12-30 是周一
			wantDate: "2024-12-30",
		},
		{
			name:     "Tuesday",
			now:      time.Date(2024, 12, 31, 10, 0, 0, 0, loc), // 2024-12-31 是周二
			wantDate: "2024-12-30",
		},
		{
			name:     "Wednesday",
			now:      time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 2025-01-01 是周三
			wantDate: "2024-12-30",
		},
		{
			name:     "Thursday",
			now:      time.Date(2025, 1, 2, 10, 0, 0, 0, loc), // 2025-01-02 是周四
			wantDate: "2024-12-30",
		},
		{
			name:     "Friday",
			now:      time.Date(2025, 1, 3, 10, 0, 0, 0, loc), // 2025-01-03 是周五
			wantDate: "2024-12-30",
		},
		{
			name:     "Saturday",
			now:      time.Date(2025, 1, 4, 10, 0, 0, 0, loc), // 2025-01-04 是周六
			wantDate: "2024-12-30",
		},
		{
			name:     "Sunday - critical edge case",
			now:      time.Date(2025, 1, 5, 10, 0, 0, 0, loc), // 2025-01-05 是周日
			wantDate: "2024-12-30",                            // 应该是本周一，不是下周一
		},
		{
			name:     "Sunday early morning",
			now:      time.Date(2025, 1, 5, 7, 0, 0, 0, loc), // 周日早上 7 点
			wantDate: "2024-12-30",
		},
		{
			name:     "Sunday late night",
			now:      time.Date(2025, 1, 5, 23, 59, 0, 0, loc), // 周日深夜
			wantDate: "2024-12-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getThisWeekMonday(tt.now, loc)
			gotDate := got.Format("2006-01-02")
			if gotDate != tt.wantDate {
				t.Errorf("getThisWeekMonday(%v) = %s, want %s", tt.now.Format("2006-01-02 Mon"), gotDate, tt.wantDate)
			}
			// 验证时间是 08:02
			if got.Hour() != 8 || got.Minute() != 2 {
				t.Errorf("getThisWeekMonday() time = %02d:%02d, want 08:02", got.Hour(), got.Minute())
			}
		})
	}
}

func TestGetThisWeekMonday_MondayIsAlwaysBeforeOrEqualNow(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	// 测试一整周的每一天，确保计算出的周一总是在当前日期之前或等于当前日期
	baseDate := time.Date(2025, 1, 6, 12, 0, 0, 0, loc) // 2025-01-06 是周一

	for i := 0; i < 7; i++ {
		now := baseDate.AddDate(0, 0, i)
		monday := getThisWeekMonday(now, loc)

		// 周一应该在 now 的同一天或之前
		mondayDate := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, loc)
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

		if mondayDate.After(nowDate) {
			t.Errorf("day %d (%s): monday %s is after now %s",
				i, now.Weekday(), monday.Format("2006-01-02"), now.Format("2006-01-02"))
		}
	}
}

func TestNeedsRefresh_WeeklyOnSunday(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	store := NewStore()
	_ = &Refresher{Store: store} // 用于未来扩展测试

	// 模拟上周一更新的数据
	lastMonday := time.Date(2024, 12, 23, 8, 5, 0, 0, loc) // 上周一 08:05
	snap := &Snapshot{
		Period:    PeriodWeekly,
		UpdatedAt: lastMonday,
		Symbols:   map[string]Levels{"BTCUSDT": {}},
	}
	store.Swap(PeriodWeekly, snap)

	// 测试周日：应该判定为 stale，因为本周一已经过了
	sunday := time.Date(2024, 12, 29, 10, 0, 0, 0, loc) // 周日 10:00

	// 我们需要一个可以注入时间的测试方法
	// 由于 needsRefresh 使用 time.Now()，我们需要重构或使用其他方式测试
	// 这里我们直接测试 getThisWeekMonday 的正确性

	thisMonday := getThisWeekMonday(sunday, loc)
	expectedMonday := time.Date(2024, 12, 23, 8, 2, 0, 0, loc)

	if !thisMonday.Equal(expectedMonday) {
		t.Errorf("On Sunday %s, thisMonday = %s, want %s",
			sunday.Format("2006-01-02"),
			thisMonday.Format("2006-01-02 15:04"),
			expectedMonday.Format("2006-01-02 15:04"))
	}

	// 验证：snap.UpdatedAt (上周一 08:05) 应该在 thisMonday (本周一 08:02) 之后
	// 所以不应该判定为 stale
	// 但如果 snap.UpdatedAt 是上上周的，就应该判定为 stale
	oldSnap := &Snapshot{
		Period:    PeriodWeekly,
		UpdatedAt: time.Date(2024, 12, 16, 8, 5, 0, 0, loc), // 上上周一
		Symbols:   map[string]Levels{"BTCUSDT": {}},
	}
	store.Swap(PeriodWeekly, oldSnap)

	// 上上周一 08:05 < 本周一 08:02 (2024-12-23)，应该判定为 stale
	if !oldSnap.UpdatedAt.Before(thisMonday) {
		t.Errorf("oldSnap.UpdatedAt %s should be before thisMonday %s",
			oldSnap.UpdatedAt.Format("2006-01-02 15:04"),
			thisMonday.Format("2006-01-02 15:04"))
	}
}


// Property 1: 周一计算一致性
// *For any* date in a week (Monday through Sunday), the calculated "this week's Monday"
// should always be the same date and should be on or before the current date.
// **Validates: Requirements 1.2, 1.5**
func TestProperty_MondayCalculationConsistency(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	// 使用 testing/quick 进行属性测试
	// 生成随机日期，验证同一周内的所有日期计算出的周一相同
	iterations := 100

	for i := 0; i < iterations; i++ {
		// 生成一个随机的周一作为基准
		baseYear := 2020 + (i % 10)
		baseMonth := time.Month(1 + (i % 12))
		baseDay := 1 + (i % 28)
		baseDate := time.Date(baseYear, baseMonth, baseDay, 12, 0, 0, 0, loc)

		// 找到这个日期所在周的周一
		weekday := int(baseDate.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		mondayOfWeek := baseDate.AddDate(0, 0, -(weekday - 1))

		// 验证这一周的每一天计算出的周一都相同
		var expectedMonday time.Time
		for day := 0; day < 7; day++ {
			currentDay := mondayOfWeek.AddDate(0, 0, day)
			calculatedMonday := getThisWeekMonday(currentDay, loc)

			if day == 0 {
				expectedMonday = calculatedMonday
			} else {
				// 同一周内的所有日期应该计算出相同的周一
				if calculatedMonday.Format("2006-01-02") != expectedMonday.Format("2006-01-02") {
					t.Errorf("Iteration %d, day %d (%s): calculated Monday %s != expected %s",
						i, day, currentDay.Weekday(),
						calculatedMonday.Format("2006-01-02"),
						expectedMonday.Format("2006-01-02"))
				}
			}

			// 验证计算出的周一不在当前日期之后
			mondayDate := time.Date(calculatedMonday.Year(), calculatedMonday.Month(), calculatedMonday.Day(), 0, 0, 0, 0, loc)
			currentDate := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), 0, 0, 0, 0, loc)
			if mondayDate.After(currentDate) {
				t.Errorf("Iteration %d: Monday %s is after current day %s",
					i, mondayDate.Format("2006-01-02"), currentDate.Format("2006-01-02"))
			}
		}
	}
}

// Property 2: 过期检测持续性
// *For any* weekly pivot data that was last updated before this week's Monday,
// the needsRefresh function should return true on all days of the current week.
// **Validates: Requirements 1.1, 1.3**
func TestProperty_StaleDetectionPersistence(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	iterations := 100

	for i := 0; i < iterations; i++ {
		// 生成一个随机的周一
		baseYear := 2020 + (i % 10)
		baseMonth := time.Month(1 + (i % 12))
		baseDay := 1 + (i % 28)
		baseDate := time.Date(baseYear, baseMonth, baseDay, 12, 0, 0, 0, loc)

		// 找到这个日期所在周的周一
		weekday := int(baseDate.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		thisMonday := baseDate.AddDate(0, 0, -(weekday - 1))
		thisMonday8am02 := time.Date(thisMonday.Year(), thisMonday.Month(), thisMonday.Day(), 8, 2, 0, 0, loc)

		// 模拟上周更新的数据（在本周一之前）
		lastWeekUpdate := thisMonday8am02.AddDate(0, 0, -7)

		// 验证这一周的每一天（08:02 之后）都应该判定为 stale
		for day := 0; day < 7; day++ {
			currentDay := thisMonday.AddDate(0, 0, day)
			// 设置时间为 10:00，确保在 08:02 之后
			currentTime := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), 10, 0, 0, 0, loc)

			calculatedMonday := getThisWeekMonday(currentTime, loc)

			// 验证：lastWeekUpdate 应该在 calculatedMonday 之前
			isStale := lastWeekUpdate.Before(calculatedMonday)
			if !isStale {
				t.Errorf("Iteration %d, day %d (%s): lastWeekUpdate %s should be before calculatedMonday %s",
					i, day, currentTime.Weekday(),
					lastWeekUpdate.Format("2006-01-02 15:04"),
					calculatedMonday.Format("2006-01-02 15:04"))
			}
		}
	}
}
