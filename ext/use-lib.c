/*
    go build -buildmode=c-shared -o libsuseconnect.so ext/main.go
    gcc ext/use-lib.c -o use-lib -L. -lsuseconnect
    LD_LIBRARY_PATH=. ./use-lib
*/

#include <stdio.h>
#include <stdlib.h>
#include "../libsuseconnect.h"

void main(void) {
    char *format = "json";
    char *statuses;

    statuses = getstatus(format);
    printf("%s\n", statuses);
    free(statuses);
}
