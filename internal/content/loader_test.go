package content

import "testing"

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    Format
		wantErr bool
	}{
		{name: "markdown", path: "notes.md", want: FormatMarkdown},
		{name: "markdown-long-ext", path: "notes.markdown", want: FormatMarkdown},
		{name: "html", path: "notes.html", want: FormatHTML},
		{name: "htm", path: "notes.htm", want: FormatHTML},
		{name: "unsupported", path: "notes.txt", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectFormat(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected format %q, got %q", tt.want, got)
			}
		})
	}
}
