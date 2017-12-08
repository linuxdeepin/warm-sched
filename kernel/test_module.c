#include <linux/module.h>
#include <linux/dcache.h>
#include <linux/kallsyms.h>
#include <linux/pagemap.h>
#include <linux/slab.h>
#include <linux/fs.h>
#include <linux/mm.h>
#include <linux/list_lru.h>
#include <linux/time.h>


static bool dump_inode(struct inode *inode);
static void traver_sb(struct super_block *sb, void *user_data);
static void show_time_elapse(struct timespec begin);

static bool SHOW = false;


static int test_module_init(void)
{
  struct file_system_type* t = get_fs_type("ext4");

  printk("Test snyh haha Module Installed\n");

  iterate_supers_type(t, traver_sb, 0);
  //put_filesystem(t);
  return 0;
}

enum lru_status inode_lru_walk (struct list_head *item, struct list_lru_one *list,
                                spinlock_t *lock, void *cb_arg)
{
  unsigned long *count = cb_arg;
  struct inode* inode = list_entry(item, struct inode, i_lru);
  if (dump_inode(inode)) (*count)++;
  return LRU_SKIP;
}

void collect_inode_in_lru(struct super_block *sb, unsigned long* count)
{
  unsigned long n = 10000000;
  list_lru_walk(&(sb->s_inode_lru), inode_lru_walk, count, n);
}

static void traver_sb(struct super_block *sb, void *user_data)
{
  struct timespec begin = CURRENT_TIME;
  struct inode *inode = NULL;
  unsigned long n1 = 0, n2 = 0;

  spin_lock(&sb->s_inode_list_lock);
  list_for_each_entry(inode, &sb->s_inodes, i_sb_list) {
    if (dump_inode(inode)) n1++;
  }
  spin_unlock(&sb->s_inode_list_lock);
  collect_inode_in_lru(sb, &n2);

  printk("For %s: %ld %ld", sb->s_id, n1, n2);
  show_time_elapse(begin);
}

static void show_time_elapse(struct timespec begin)
{
  struct timespec now = CURRENT_TIME;
  printk("Elapse %lld.%.9lds\n", (long long)(now.tv_sec - begin.tv_sec), now.tv_nsec - begin.tv_nsec);
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

static bool dump_inode(struct inode *inode)
{
  struct dentry *d = NULL;
  sector_t bn = 0;
  static char bufname[1024];

  d = d_find_any_alias(inode);
  if (d == 0) {
    return false;
  }

  spin_lock(&inode->i_lock);
  if (!S_ISREG(inode->i_mode) || inode->i_mapping->nrpages == 0) {
    spin_unlock(&inode->i_lock);
    dput(d);
    return false;
  }

  /* spin_unlock(&inode->i_lock); */
  /* bn = bmap(inode, 0); */
  /* spin_lock(&inode->i_lock); */

  if (SHOW)
    printk("Inode:%ld\t%ld\t%lld\t%s\t%ld\n", inode->i_ino,
           inode->i_mapping->nrpages,
           ((((loff_t)inode->i_blocks) << 9) + inode->i_bytes) / 4096,
           dentry_path_raw(d, bufname, sizeof(bufname)),
           bn
           );
  dput(d);
  spin_unlock(&inode->i_lock);
  return true;
}
