#include <linux/module.h>
#include <linux/dcache.h>
#include <linux/kallsyms.h>
#include <linux/pagemap.h>
#include <linux/slab.h>
#include	<linux/mm.h>

static void is_dentry_in_cache(void);


static int test_module_init(void)
{
  printk("Test snyh haha Module Installed\n");
  is_dentry_in_cache();
  return 0;
}


static void is_dentry_in_cache(void)
{
  char *sym_name = "dentry_cache";
  //  kmem_cache* unsafe_dentry_cache = kallsyms_lookup_name(sym_name);
  //  printk("Test found dentry:%p\n", unsafe_dentry_cache);
}

static void test_module_exit(void)
{
  printk("Test Module Removed\n");
}

module_init(test_module_init);
module_exit(test_module_exit);

MODULE_AUTHOR("IsonProjects");
MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("Simple Kernel Module");
