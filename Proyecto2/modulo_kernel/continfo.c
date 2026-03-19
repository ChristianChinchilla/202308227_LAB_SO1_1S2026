#include <linux/module.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/mm.h>
#include <linux/sched/signal.h>

#define PROCFS_NAME "continfo_pr2_so1_202308227"

MODULE_LICENSE("GPL");
MODULE_AUTHOR("202308227");

static int my_proc_show(struct seq_file *m, void *v) {
    struct sysinfo i;
    struct task_struct *task;
    unsigned long total_ram, free_ram, used_ram;
    unsigned long vsz, rss, cpu_time;
    bool first_process = true;

    si_meminfo(&i);
    total_ram = (i.totalram * i.mem_unit) / (1024 * 1024);
    free_ram = (i.freeram * i.mem_unit) / (1024 * 1024);
    used_ram = total_ram - free_ram;

    seq_printf(m, "{\n  \"ram\": {\"total\": %lu, \"free\": %lu, \"used\": %lu},\n", total_ram, free_ram, used_ram);
    seq_printf(m, "  \"procesos\": [\n");

    for_each_process(task) {
        vsz = rss = 0;
        cpu_time = task->utime + task->stime;
        if (task->mm) {
            vsz = task->mm->total_vm << (PAGE_SHIFT - 10);
            rss = get_mm_rss(task->mm) << (PAGE_SHIFT - 10);
        }
        if (!first_process) seq_printf(m, ",\n");
        seq_printf(m, "    {\"pid\": %d, \"nombre\": \"%s\", \"vsz\": %lu, \"rss\": %lu, \"cpu\": %lu}",
                   task->pid, task->comm, vsz, rss, cpu_time);
        first_process = false;
    }
    seq_printf(m, "\n  ]\n}\n");
    return 0;
}

static int my_proc_open(struct inode *inode, struct file *file) { return single_open(file, my_proc_show, NULL); }

static const struct proc_ops my_proc_fops = {
    .proc_open = my_proc_open, .proc_read = seq_read, .proc_lseek = seq_lseek, .proc_release = single_release,
};

static int __init my_init(void) { proc_create(PROCFS_NAME, 0, NULL, &my_proc_fops); return 0; }
static void __exit my_exit(void) { remove_proc_entry(PROCFS_NAME, NULL); }

module_init(my_init);
module_exit(my_exit);
