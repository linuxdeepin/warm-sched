#include <linux/module.h>
#include <linux/dcache.h>

static int test_module_init(void)
{
  printk("Test Module Installed\n");

  return 0;
}

void is_dentry_in_cache(void)
{
  struct qstr str = {
    .hash_len = 5,
    .name = "hehe",
  };
  d_lookup(0, &str);
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
