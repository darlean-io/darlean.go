copy embedlib\embedlib.dll examples\embed-c
copy embedlib\embedlib.h examples\embed-c
cd examples/embed-c
gcc -o test test.c embedlib.dll
cd ../..