#include <linux/module.h>
#include <linux/kernel.h>

MODULE_LICENSE("GPL");
MODULE_AUTHOR("202308227");

int init_module(void) {
    printk(KERN_INFO "Hola Mundo 202308227 (desde Rust-C Wrapper)\n");
    return 0;
}

void cleanup_module(void) {
    printk(KERN_INFO "Adios Mundo 202308227\n");
}
