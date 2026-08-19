package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSearchInvitationsFiltersAndPaginates(t *testing.T) {
	require.NoError(t, DB.AutoMigrate(&Invitation{}))
	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Invitation{}).Error)
	t.Cleanup(func() {
		require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Invitation{}).Error)
	})

	now := common.GetTimestamp()
	invitations := []Invitation{
		{Id: 1, Name: "alpha-active", Key: "INV-000000000000000000000000000001", Status: common.InvitationCodeStatusEnabled, Quota: 100, Group: "default", ExpiredTime: 0},
		{Id: 2, Name: "alpha-future", Key: "INV-000000000000000000000000000002", Status: common.InvitationCodeStatusEnabled, Quota: 500, Group: "vip", ExpiredTime: now + 3600},
		{Id: 3, Name: "alpha-expired", Key: "INV-000000000000000000000000000003", Status: common.InvitationCodeStatusEnabled, Quota: 100, Group: "default", ExpiredTime: now - 10},
		{Id: 4, Name: "beta-disabled", Key: "INV-000000000000000000000000000004", Status: common.InvitationCodeStatusDisabled, Quota: 100, Group: "default", ExpiredTime: 0},
		{Id: 5, Name: "beta-used", Key: "INV-000000000000000000000000000005", Status: common.InvitationCodeStatusUsed, Quota: 100, Group: "default", ExpiredTime: 0},
	}
	require.NoError(t, DB.Create(&invitations).Error)

	tests := []struct {
		name      string
		keyword   string
		status    string
		startIdx  int
		num       int
		wantTotal int64
		wantIds   []int
	}{
		{
			name:      "no filters returns all rows",
			num:       10,
			wantTotal: 5,
			wantIds:   []int{5, 4, 3, 2, 1},
		},
		{
			name:      "keyword filters by name prefix",
			keyword:   "alpha",
			num:       10,
			wantTotal: 3,
			wantIds:   []int{3, 2, 1},
		},
		{
			name:      "enabled status excludes expired rows",
			status:    "1",
			num:       10,
			wantTotal: 2,
			wantIds:   []int{2, 1},
		},
		{
			name:      "expired status returns enabled expired rows",
			status:    "expired",
			num:       10,
			wantTotal: 1,
			wantIds:   []int{3},
		},
		{
			name:      "disabled status",
			status:    "2",
			num:       10,
			wantTotal: 1,
			wantIds:   []int{4},
		},
		{
			name:      "used status",
			status:    "3",
			num:       10,
			wantTotal: 1,
			wantIds:   []int{5},
		},
		{
			name:      "pagination keeps unpaged total",
			startIdx:  1,
			num:       2,
			wantTotal: 5,
			wantIds:   []int{4, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, total, err := SearchInvitations(tt.keyword, tt.status, tt.startIdx, tt.num)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			gotIds := make([]int, 0, len(rows))
			for _, row := range rows {
				gotIds = append(gotIds, row.Id)
			}
			assert.Equal(t, tt.wantIds, gotIds)
		})
	}
}

func TestDeleteInvalidInvitations(t *testing.T) {
	require.NoError(t, DB.AutoMigrate(&Invitation{}))
	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Invitation{}).Error)
	t.Cleanup(func() {
		require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Invitation{}).Error)
	})

	now := common.GetTimestamp()
	invitations := []Invitation{
		{Id: 1, Name: "keep-valid", Key: "INV-K1", Status: common.InvitationCodeStatusEnabled, ExpiredTime: 0},
		{Id: 2, Name: "keep-future", Key: "INV-K2", Status: common.InvitationCodeStatusEnabled, ExpiredTime: now + 3600},
		{Id: 3, Name: "del-expired", Key: "INV-D1", Status: common.InvitationCodeStatusEnabled, ExpiredTime: now - 10},
		{Id: 4, Name: "del-disabled", Key: "INV-D2", Status: common.InvitationCodeStatusDisabled, ExpiredTime: 0},
		{Id: 5, Name: "del-used", Key: "INV-D3", Status: common.InvitationCodeStatusUsed, ExpiredTime: 0},
	}
	require.NoError(t, DB.Create(&invitations).Error)

	deleted, err := DeleteInvalidInvitations()
	require.NoError(t, err)
	assert.Equal(t, int64(3), deleted)

	var remaining []Invitation
	require.NoError(t, DB.Order("id asc").Find(&remaining).Error)
	require.Len(t, remaining, 2)
	assert.Equal(t, 1, remaining[0].Id)
	assert.Equal(t, 2, remaining[1].Id)
}
