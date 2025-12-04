package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRoleLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    RoleLevel
		wantErr bool
	}{
		{
			name:    "valid admin role",
			input:   "admin",
			want:    RoleLevelAdmin,
			wantErr: false,
		},
		{
			name:    "valid editor role",
			input:   "editor",
			want:    RoleLevelEditor,
			wantErr: false,
		},
		{
			name:    "valid viewer role",
			input:   "viewer",
			want:    RoleLevelViewer,
			wantErr: false,
		},
		{
			name:    "case insensitive - Admin",
			input:   "Admin",
			want:    RoleLevelAdmin,
			wantErr: false,
		},
		{
			name:    "case insensitive - EDITOR",
			input:   "EDITOR",
			want:    RoleLevelEditor,
			wantErr: false,
		},
		{
			name:    "invalid role",
			input:   "superuser",
			wantErr: true,
		},
		{
			name:    "empty role",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRoleLevel(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid role level")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRoleLevel_String(t *testing.T) {
	tests := []struct {
		role     RoleLevel
		expected string
	}{
		{RoleLevelAdmin, "admin"},
		{RoleLevelEditor, "editor"},
		{RoleLevelViewer, "viewer"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.String())
		})
	}
}

func TestRoleLevel_CanRead(t *testing.T) {
	assert.True(t, RoleLevelAdmin.CanRead(), "admin should have read permission")
	assert.True(t, RoleLevelEditor.CanRead(), "editor should have read permission")
	assert.True(t, RoleLevelViewer.CanRead(), "viewer should have read permission")
}

func TestRoleLevel_CanWrite(t *testing.T) {
	assert.True(t, RoleLevelAdmin.CanWrite(), "admin should have write permission")
	assert.True(t, RoleLevelEditor.CanWrite(), "editor should have write permission")
	assert.False(t, RoleLevelViewer.CanWrite(), "viewer should NOT have write permission")
}

func TestRoleLevel_CanAdmin(t *testing.T) {
	assert.True(t, RoleLevelAdmin.CanAdmin(), "admin should have admin permission")
	assert.False(t, RoleLevelEditor.CanAdmin(), "editor should NOT have admin permission")
	assert.False(t, RoleLevelViewer.CanAdmin(), "viewer should NOT have admin permission")
}

func TestRoleLevel_Hierarchy(t *testing.T) {
	// Test permission hierarchy: Admin > Editor > Viewer
	
	// Admin has all permissions
	assert.True(t, RoleLevelAdmin.CanRead())
	assert.True(t, RoleLevelAdmin.CanWrite())
	assert.True(t, RoleLevelAdmin.CanAdmin())
	
	// Editor has read and write
	assert.True(t, RoleLevelEditor.CanRead())
	assert.True(t, RoleLevelEditor.CanWrite())
	assert.False(t, RoleLevelEditor.CanAdmin())
	
	// Viewer has only read
	assert.True(t, RoleLevelViewer.CanRead())
	assert.False(t, RoleLevelViewer.CanWrite())
	assert.False(t, RoleLevelViewer.CanAdmin())
}

