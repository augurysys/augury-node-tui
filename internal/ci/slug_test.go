package ci

import "testing"

func TestSlugFromRemote(t *testing.T) {
	tests := []struct {
		name    string
		remote  string
		want    string
		wantErr bool
	}{
		{
			name:   "SSH format",
			remote: "git@github.com:augurysys/augury-node.git",
			want:   "gh/augurysys/augury-node",
		},
		{
			name:   "HTTPS format",
			remote: "https://github.com/augurysys/augury-node.git",
			want:   "gh/augurysys/augury-node",
		},
		{
			name:   "HTTPS without .git",
			remote: "https://github.com/augurysys/augury-node",
			want:   "gh/augurysys/augury-node",
		},
		{
			name:   "SSH without .git",
			remote: "git@github.com:augurysys/augury-node",
			want:   "gh/augurysys/augury-node",
		},
		{
			name:    "unsupported",
			remote:  "https://gitlab.com/org/repo.git",
			wantErr: true,
		},
		{
			name:    "empty",
			remote:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SlugFromRemote(tt.remote)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
