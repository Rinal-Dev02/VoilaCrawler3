package s3

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestS3Client_Put(t *testing.T) {
	f, err := os.OpenFile("s3.go", os.O_RDONLY, 644)
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		host       string
		bucket     string
		httpClient *http.Client
	}
	type args struct {
		ctx    context.Context
		name   string
		reader io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object
		wantErr bool
	}{
		{
			name: "upload",
			fields: fields{
				bucket:     "voila-downloads",
				host:       "172.31.130.253:32389",
				httpClient: &http.Client{},
			},
			args: args{
				ctx:    context.Background(),
				name:   "s3.go",
				reader: f,
			},
			want: &object{
				Name:   "s3.go",
				Scheme: "https",
				Domain: "s3.voila.love",
				Path:   "/paas/s3/object/voila-downloads/s3.go",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &S3Client{
				host:       tt.fields.host,
				bucket:     tt.fields.bucket,
				httpClient: tt.fields.httpClient,
			}
			got, err := c.Put(tt.args.ctx, tt.args.name, tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("S3Client.Put() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("S3Client.Put() = %v", got)
		})
	}
}
