typedef void (*logLineFunc)(int level, const char *message);

// this function enables calling of FFI callback function passed
// to Go via function pointer
void log_bridge_fun(logLineFunc f, int level, const char *message)
{
    f(level, message);
}
