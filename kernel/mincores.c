#include <linux/module.h>
#include <linux/fs.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/fs_struct.h>
#include <linux/mount.h>
#include <linux/sched.h>
#include <linux/version.h>

static const char* PROC_NAME = "mincores";

static void dump_mapping(struct seq_file*sf, unsigned long total, struct address_space* addr);
static bool dump_inode(struct seq_file*, struct inode *i);
static void traver_sb(struct seq_file*, struct super_block *sb, void *user_data);


#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 1, 0)

#define RADIX_TREE_EXCEPTIONAL_ENTRY	2

#define radix_tree_for_each_contig(slot, root, iter, start)		\
	for (slot = radix_tree_iter_init(iter, start) ;			\
	     slot || (slot = radix_tree_next_chunk(root, iter,		\
				RADIX_TREE_ITER_CONTIG)) ;		\
	     slot = radix_tree_next_slot(slot, iter,			\
				RADIX_TREE_ITER_CONTIG))

static inline int radix_tree_exceptional_entry(void *arg)
{
	/* Not unlikely because radix_tree_exception often tested first */
	return (unsigned long)arg & RADIX_TREE_EXCEPTIONAL_ENTRY;
}

#endif

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

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 7, 7)
static const struct proc_ops proc_file_fops = {
	.proc_read	= seq_read,
	.proc_open	= proc_open_callback,
	.proc_release	= single_release,
	.proc_lseek	= seq_lseek,
};
#else
static const struct file_operations proc_file_fops = {
    .owner = THIS_MODULE,
    .open = proc_open_callback,
    .read	= seq_read,
    .llseek = seq_lseek,
    .release = single_release,
};
#endif

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
    if (inode->i_nlink == 0 || !S_ISREG(inode->i_mode) || inode->i_mapping->nrpages == 0) {
        spin_unlock(&inode->i_lock);
        return true;
    }
    spin_unlock(&inode->i_lock);
    return false;
}


#if LINUX_VERSION_CODE <= KERNEL_VERSION(4, 17, 0)
#define i_pages page_tree
#else
#define i_pages i_pages
#endif

static void dump_mapping(struct seq_file*sf, unsigned long total, struct address_space* addr)
{
    void **slot;
    struct radix_tree_iter iter;

    unsigned long start=0, end = 0, next_start = 0;
    bool found;

#ifdef DEBUG_MAPPING
    unsigned long debug1=0, debug2 = 0;
#endif

    do {
        found = false;
        rcu_read_lock();
        radix_tree_for_each_contig(slot, &addr->i_pages, &iter, next_start) {
            if (0 != radix_tree_exceptional_entry(radix_tree_deref_slot(slot))) {
                break;
            }
            end = iter.index;
            found = true;

#ifdef DEBUG_MAPPING
            debug1++;
            if (end > iter.index || iter.index < start) {
                seq_printf(sf, "ERROR0: %ld < %ld < %ld ", start, iter.index, end);
            }
#endif
        }
        rcu_read_unlock();
        if (found) {
#ifdef DEBUG_MAPPING
            debug2 += (end - start + 1);
#endif
            seq_printf(sf, "[%ld:%ld],", start, end);
            start = end;
            next_start = end+1;
        } else {
            next_start++;
            start = next_start;
        }
    } while (found || next_start < total);

#ifdef DEBUG_MAPPING
    if (debug1 != addr->nrpages) {
        seq_printf(sf, " ERROR2 (%ld != %ld) ", debug1, addr->nrpages);
    }
    if (debug2 != addr->nrpages) {
        seq_printf(sf, " ERROR3 (%ld != %ld) ", debug2, addr->nrpages);
    }
#endif
}

static bool dump_inode(struct seq_file* sf, struct inode *inode)
{
    struct dentry *d = NULL;
    static char bufname[1024];
    char* tmpname = 0;
    
#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 7, 7)
    sector_t block = 0;
    int bn = 0;
#else
    sector_t bn = 0;
#endif
    loff_t fs = i_size_read(inode);
    unsigned long nrpages = inode->i_mapping->nrpages;
    unsigned long total = (fs + PAGE_SIZE - 1) / PAGE_SIZE;
    if (fs == 0) {
        return false;
    }
    if (total < nrpages) {
        // Don't known why if the file size is 20480, this will be happened.
        total = nrpages;
    }

    if (skip_inode(inode)) {
        return false;
    }

    d = d_find_any_alias(inode);
    if (d == 0) {
        return false;
    }
    tmpname = dentry_path_raw(d, bufname, sizeof(bufname));
    dput(d);

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 7, 7)
    bn = bmap(inode, &block);
#else
    bn = bmap(inode, 0);
#endif

    seq_printf(sf, "%lld\t%ld\t", bn, total);

    spin_lock(&inode->i_lock);
    dump_mapping(sf, total, inode->i_mapping);
    spin_unlock(&inode->i_lock);

    seq_printf(sf, "\t%s\n", tmpname);
    return true;
}
