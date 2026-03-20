#include <linux/module.h>
#define INCLUDE_VERMAGIC
#include <linux/build-salt.h>
#include <linux/elfnote-lto.h>
#include <linux/export-internal.h>
#include <linux/vermagic.h>
#include <linux/compiler.h>

#ifdef CONFIG_UNWINDER_ORC
#include <asm/orc_header.h>
ORC_HEADER;
#endif

BUILD_SALT;
BUILD_LTO_INFO;

MODULE_INFO(vermagic, VERMAGIC_STRING);
MODULE_INFO(name, KBUILD_MODNAME);

__visible struct module __this_module
__section(".gnu.linkonce.this_module") = {
	.name = KBUILD_MODNAME,
	.init = init_module,
#ifdef CONFIG_MODULE_UNLOAD
	.exit = cleanup_module,
#endif
	.arch = MODULE_ARCH_INIT,
};

#ifdef CONFIG_RETPOLINE
MODULE_INFO(retpoline, "Y");
#endif



static const char ____versions[]
__used __section("__versions") =
	"\x14\x00\x00\x00\x66\x12\xef\xfc"
	"proc_create\0"
	"\x10\x00\x00\x00\x7e\x3a\x2c\x12"
	"_printk\0"
	"\x14\x00\x00\x00\x5a\x06\x55\xcf"
	"single_open\0"
	"\x14\x00\x00\x00\x21\xb0\x5d\x09"
	"init_task\0\0\0"
	"\x10\x00\x00\x00\xad\x64\xb7\xdc"
	"memset\0\0"
	"\x14\x00\x00\x00\x7c\x24\xc7\x40"
	"si_meminfo\0\0"
	"\x14\x00\x00\x00\xd9\x68\x1e\x0d"
	"seq_printf\0\0"
	"\x1c\x00\x00\x00\xcb\xf6\xfd\xf0"
	"__stack_chk_fail\0\0\0\0"
	"\x1c\x00\x00\x00\x22\x84\x32\x75"
	"remove_proc_entry\0\0\0"
	"\x14\x00\x00\x00\xc9\x36\xc6\x42"
	"seq_read\0\0\0\0"
	"\x14\x00\x00\x00\x55\xc8\x0f\x78"
	"seq_lseek\0\0\0"
	"\x18\x00\x00\x00\x13\xdb\x51\x1e"
	"single_release\0\0"
	"\x18\x00\x00\x00\xd4\xce\x9f\xb7"
	"module_layout\0\0\0"
	"\x00\x00\x00\x00\x00\x00\x00\x00";

MODULE_INFO(depends, "");


MODULE_INFO(srcversion, "FA045F29223D198F4FC1243");
