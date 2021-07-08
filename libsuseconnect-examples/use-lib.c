/*
    cd ..
    make build-so-example
*/

#include <stdio.h>
#include <stdlib.h>
#include "../out/libsuseconnect.h"

void main(void) {
    char *format = "json";
    char *statuses;

    statuses = getstatus(format);
    printf("%s\n", statuses);
    free(statuses);
}
