#include <linux/module.h>
#include <linux/dcache.h>
#include <linux/fs.h>
#include <linux/mm.h>
#include <linux/time.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/sched.h>
#include <linux/fs_struct.h>
#include <linux/mount.h>
#include <linux/backing-dev-defs.h>
#include <linux/genhd.h>

static bool dump_inode(struct inode *i, struct vfsmount* mnt, const char*);
static void traver_sb(struct super_block *sb, void *user_data);
static void show_time_elapse(struct timespec begin);

static struct proc_dir_entry *proc_file_entry;
static struct seq_file* SF = 0;

static bool is_normal_fs_type(struct vfsmount *mnt)
{
  const char* typ = 0;
  if (mnt == 0 || mnt->mnt_sb == 0 || mnt->mnt_root == 0) {
    return false;
  }
  typ = mnt->mnt_sb->s_type->name;
  if (strcmp(typ, "ext3") == 0||
      strcmp(typ, "ext4") == 0||
      strcmp(typ, "ext2") == 0||
      strcmp(typ, "fat") == 0||
      strcmp(typ, "ntfs") == 0) {
    return true;
  }
  return false;
}

static int traver_vfsmount(struct vfsmount * mnt, void *user_data)
{
  struct super_block *sb = mnt->mnt_sb;

  if (mnt->mnt_root == 0) {
    // TODO: why the mnt->mnt_sb could be a invalid pointer link 0000003d when mnt->mnt_root==0
    seq_printf(SF, "BAD SB %p %p\n", mnt, sb);
    return 0;
  }

  if (is_normal_fs_type(mnt)) {
    traver_sb(sb, mnt);
  }
  return 0;
}

static int show_proc_content(struct seq_file *filp, void *p)
{
  struct path root;
  SF = filp;

  get_fs_root(current->fs, &root);
  iterate_mounts(traver_vfsmount, 0, root.mnt);
  path_put(&root);

  SF = 0;
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

static void traver_sb(struct super_block *sb, void *user_data)
{
  struct vfsmount* mnt = user_data;
  struct timespec begin = CURRENT_TIME;
  struct inode *i = NULL;
  unsigned long n1 = 0;

  char bname[1024];
  bdevname(sb->s_bdev, bname);

  spin_lock(&sb->s_inode_list_lock);
  list_for_each_entry(i, &sb->s_inodes, i_sb_list) {
    if (dump_inode(i, mnt, bname)) n1++;
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


static bool skip_inode(struct inode* inode)
{
  spin_lock(&inode->i_lock);
  if (!S_ISREG(inode->i_mode) || inode->i_mapping->nrpages == 0) {
    spin_unlock(&inode->i_lock);
    return true;
  }
  spin_unlock(&inode->i_lock);
  return false;
}

static bool dump_inode(struct inode *inode, struct vfsmount* mnt, const char* dev_name)
{
  struct dentry *d = NULL;
  static char bufname[1024];
  sector_t bn = 0;
  size_t ms = 0;
  loff_t fs = inode_get_bytes(inode);
  if (fs == 0) {
    return false;
  }

  if (skip_inode(inode)) {
    return false;
  }

  d = d_find_any_alias(inode);
  if (d == 0) {
    return false;
  }

  bn = bmap(inode, 0);

  spin_lock(&inode->i_lock);
  ms = inode->i_mapping->nrpages * PAGE_SIZE;
  if (!S_ISREG(inode->i_mode) || inode->i_mapping->nrpages == 0) {
    spin_unlock(&inode->i_lock);
    dput(d);
    return false;
  }

  if (SF) {
    struct path r = {
      .dentry = d,
      .mnt = mnt,
    };
    seq_printf(SF, "%zuKB\t%llu%%\t%s:%ld\t%s\n",
               ms / 1024,
               (100 * ms / fs),
               dev_name, bn,
               d_path(&r, bufname, sizeof(bufname))
               );
  }
  dput(d);
  spin_unlock(&inode->i_lock);
  return true;
}
