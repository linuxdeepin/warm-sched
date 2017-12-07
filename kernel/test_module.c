#include <linux/module.h>
#include <linux/dcache.h>
#include <linux/kallsyms.h>
#include <linux/pagemap.h>
#include <linux/slab.h>
#include <linux/fs.h>
#include <linux/mm.h>
#include <linux/list_lru.h>

static void is_dentry_in_cache(void);


static int test_module_init(void)
{
  printk("Test snyh haha Module Installed\n");
  is_dentry_in_cache();
  return 0;
}

void traver_fn2(struct super_block *sb, void *user_data)
{
  struct inode *inode = NULL;
  struct dentry *d = NULL;
    unsigned long n = 0;
    static char bufname[1024];
    spin_lock(&sb->s_inode_list_lock);
    list_for_each_entry(inode, &sb->s_inodes, i_sb_list) {
        d = d_find_any_alias(inode);
        if (d == 0) {
          continue;
        }

        spin_lock(&inode->i_lock);

        if (!S_ISREG(inode->i_mode) || inode->i_mapping->nrpages == 0) {
          spin_unlock(&inode->i_lock);
          dput(d);
          continue;
        }
        n++;

        spin_unlock(&inode->i_lock);
        sector_t bn = bmap(inode, 0);
        spin_lock(&inode->i_lock);

        printk("Test %ld\t%ld\t%ld\t%ld\t%s\t%ld\n", n, inode->i_ino,
               inode->i_mapping->nrpages,
               (((loff_t)inode->i_blocks) << 9) + inode->i_bytes,
               dentry_path_raw(d, bufname, sizeof(bufname)),
               bn
               );
        dput(d);
        spin_unlock(&inode->i_lock);
    }
    spin_unlock(&sb->s_inode_list_lock);
}

static void is_dentry_in_cache(void)
{
  struct file_system_type* t = get_fs_type("ext4");
  iterate_supers_type(t, traver_fn2, 0);
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
