#include <linux/module.h>
#include <linux/dcache.h>
#include <linux/fs.h>
#include <linux/mm.h>
#include <linux/list_lru.h>
#include <linux/time.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>


static bool dump_inode(struct inode *inode);
static void traver_sb(struct super_block *sb, void *user_data);
static void show_time_elapse(struct timespec begin);

static struct proc_dir_entry *proc_file_entry;
static struct seq_file* SF = 0;

static int show_proc_content(struct seq_file *filp, void *p)
{
  struct file_system_type* t = get_fs_type("ext4");
  SF = filp;
  iterate_supers_type(t, traver_sb, 0);
  SF = 0;
  //put_filesystem(t);
  return 0;
}

static int proc_open_callback(struct inode *inode, struct file *filp)
{
  return single_open(filp, show_proc_content, 0);
}

static const struct file_operations proc_file_fops = {
  .owner = THIS_MODULE,
  .open = proc_open_callback,
  .read	= seq_read,
  .llseek = seq_lseek,
  .release = single_release,
};

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
  unsigned long n1 = 0;

  spin_lock(&sb->s_inode_list_lock);
  list_for_each_entry(inode, &sb->s_inodes, i_sb_list) {
    if (dump_inode(inode)) n1++;
  }
  spin_unlock(&sb->s_inode_list_lock);

  seq_printf(SF, "%s\t%ld\t", sb->s_id, n1);
  show_time_elapse(begin);
}

static void show_time_elapse(struct timespec begin)
{
  struct timespec now = CURRENT_TIME;
  seq_printf(SF, "%lld.%.9lds\n", (long long)(now.tv_sec - begin.tv_sec), now.tv_nsec - begin.tv_nsec);
}

static int test_module_init(void)
{
  printk("Test snyh haha Module Installed\n");

  proc_file_entry = proc_create("snyh123", 0, NULL, &proc_file_fops);
  if (proc_file_entry == NULL)
    return -ENOMEM;
  return 0;
}

static void test_module_exit(void)
{
  remove_proc_entry("snyh123", NULL);
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

  if (SF)
    seq_printf(SF, "%ld\t%ld\t%s\t\n",
               inode->i_mapping->nrpages,
               inode->i_ino,
               dentry_path_raw(d, bufname, sizeof(bufname))
               );
  dput(d);
  spin_unlock(&inode->i_lock);
  return true;
}
