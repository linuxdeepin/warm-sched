#include <linux/module.h>
#include <linux/fs.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/fs_struct.h>
#include <linux/mount.h>

static const char* PROC_NAME = "mincores";

static void dump_mapping(struct seq_file*sf, struct address_space* addr);
static bool dump_inode(struct seq_file*, struct inode *i);
static void traver_sb(struct seq_file*, struct super_block *sb, void *user_data);

static bool is_normal_fs_type(struct super_block* sb)
{
  const char* typ = 0;
  if (sb == 0 || sb->s_bdev == 0) {
    return false;
  }
  typ = sb->s_type->name;
  if (strcmp(typ, "ext3") == 0||
      strcmp(typ, "ext4") == 0||
      strcmp(typ, "ext2") == 0||
      strcmp(typ, "fat") == 0||
      strcmp(typ, "ntfs") == 0) {
    return true;
  }
  return false;
}

static int show_proc_content(struct seq_file *filp, void *p)
{
  struct path root;
  struct super_block *bs = 0;

  get_fs_pwd(current->fs, &root);
  bs = root.mnt->mnt_sb;
  path_put(&root);

  if (is_normal_fs_type(bs)) {
    traver_sb(filp, bs, 0);
    return 0;
  }

  return -EINVAL;
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

static void traver_sb(struct seq_file* sf, struct super_block *sb, void *user_data)
{
  struct inode *i, *ii = NULL;

  spin_lock(&sb->s_inode_list_lock);
  list_for_each_entry_safe(i, ii, &sb->s_inodes, i_sb_list) {
    spin_unlock(&sb->s_inode_list_lock);
    dump_inode(sf, i);
    spin_lock(&sb->s_inode_list_lock);
  }
  spin_unlock(&sb->s_inode_list_lock);
}

static int test_module_init(void)
{
  struct proc_dir_entry *proc_file_entry = proc_create(PROC_NAME, 0, NULL, &proc_file_fops);
  if (proc_file_entry == NULL)
    return -ENOMEM;
  return 0;
}

static void test_module_exit(void)
{
  remove_proc_entry(PROC_NAME, NULL);
}

module_init(test_module_init);
module_exit(test_module_exit);

MODULE_AUTHOR("snyh@snyh.org");
MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("Dump all file mapping info (according the PWD) from PageCache, See also mincore(2)."
                   "    Only support ext2/3/4, fat and ntfs format and only support one partition per call."
                   );

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

static void dump_mapping(struct seq_file*sf, struct address_space* addr)
{
  void **slot;
  struct radix_tree_iter iter;

  unsigned long start=0, end = 0, next_start = 0;
  bool found;

  do {
    found = false;
    radix_tree_for_each_contig(slot, &addr->page_tree, &iter, next_start) {
      end = iter.index;
      next_start = iter.next_index;
      found = true;
    }
    if (found) {
      seq_printf(sf, "[%ld:%ld],", start, end);
      start = next_start;
    }
  } while (found);
}

static bool dump_inode(struct seq_file* sf, struct inode *inode)
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

  seq_printf(sf, "%ld\t%lld\t%s\t",
             bn,
             (fs + PAGE_SIZE - 1) / PAGE_SIZE,
             dentry_path_raw(d, bufname, sizeof(bufname))
             );
  dump_mapping(sf, inode->i_mapping);
  seq_printf(sf, "\n");

  spin_unlock(&inode->i_lock);
  dput(d);

  return true;
}
