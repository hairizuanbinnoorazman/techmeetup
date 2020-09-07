package app

import "testing"

func TestBasicAuthStore_StoreMeetupToken(t *testing.T) {
	type fields struct {
		filePath string
	}
	type args struct {
		m MeetupToken
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				filePath: "../authstore.yaml",
			},
			args: args{
				m: MeetupToken{
					RefreshToken: "Aa",
					AccessToken:  "bb",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BasicAuthStore{
				filePath: tt.fields.filePath,
			}
			if err := b.StoreMeetupToken(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("BasicAuthStore.StoreMeetupToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
