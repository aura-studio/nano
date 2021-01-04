if not exist "nanoc" (cd ..)
if not exist "nanoc" (cd ..)
docker run -it --rm ^
-v %cd%:/go/src/nanox ^
golang:latest go env -w GOPROXY=https://goproxy.cn,direct && go build -o ./bin/libnanoc.so -buildmode=c-shared ./nanoc/