bin: bin/sslsync_darwin bin/sslsync_linux bin/sslsync_windows.exe

bin/sslsync_darwin:
	GOOS=darwin go build -o bin/sslsync_darwin

bin/sslsync_linux:
	GOOS=linux go build -o bin/sslsync_linux

bin/sslsync_windows.exe:
	GOOS=windows go build -o bin/sslsync_windows.exe