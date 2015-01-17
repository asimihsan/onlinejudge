#include <cerrno>
#include <cstring>
#include <exception>
#include <iostream>
#include <seccomp.h>
#include <signal.h>
#include <stdexcept>
#include <sys/prctl.h>
#include <sys/resource.h>
#include <unistd.h>
#include <unistd.h>
#include <utility>

#ifndef __x86_64__
    #error Unsupported platform
#endif

#include "seccomp-bpf.h"

void e(char const * const what) {
    throw std::runtime_error(what);
}

static int install_syscall_filter(void) {
   struct sock_filter filter[] = {
       /* Validate architecture. */
       VALIDATE_ARCHITECTURE,
       /* Grab the system call number. */
       EXAMINE_SYSCALL,
       /* List allowed syscalls. */
       ALLOW_SYSCALL(read),
       ALLOW_SYSCALL(write),
       ALLOW_SYSCALL(open),
       ALLOW_SYSCALL(close),
       ALLOW_SYSCALL(stat),
       ALLOW_SYSCALL(fstat),
       ALLOW_SYSCALL(lstat),
       ALLOW_SYSCALL(poll),
       ALLOW_SYSCALL(lseek),
       ALLOW_SYSCALL(mmap),
       ALLOW_SYSCALL(mprotect),
       ALLOW_SYSCALL(munmap),
       ALLOW_SYSCALL(brk),
       ALLOW_SYSCALL(rt_sigaction),
       ALLOW_SYSCALL(rt_sigprocmask),
       ALLOW_SYSCALL(rt_sigreturn),
       ALLOW_SYSCALL(ioctl),
       ALLOW_SYSCALL(pread64),
       ALLOW_SYSCALL(pwrite64),
       ALLOW_SYSCALL(readv),
       ALLOW_SYSCALL(writev),
       ALLOW_SYSCALL(access),
       ALLOW_SYSCALL(pipe),
       ALLOW_SYSCALL(select),
       ALLOW_SYSCALL(sched_yield),
       ALLOW_SYSCALL(mremap),
       ALLOW_SYSCALL(msync),
       ALLOW_SYSCALL(mincore),
       ALLOW_SYSCALL(madvise),
       ALLOW_SYSCALL(shmget),
       ALLOW_SYSCALL(shmat),
       ALLOW_SYSCALL(shmctl),
       ALLOW_SYSCALL(dup),
       ALLOW_SYSCALL(dup2),
       ALLOW_SYSCALL(pause),
       ALLOW_SYSCALL(nanosleep),
       ALLOW_SYSCALL(getitimer),
       ALLOW_SYSCALL(alarm),
       ALLOW_SYSCALL(setitimer),
       ALLOW_SYSCALL(getpid),
       //ALLOW_SYSCALL(sendfile),
       //ALLOW_SYSCALL(socket),
       //ALLOW_SYSCALL(connect),
       //ALLOW_SYSCALL(accept),
       //ALLOW_SYSCALL(sendto),
       //ALLOW_SYSCALL(recvfrom),
       //ALLOW_SYSCALL(sendmsg),
       //ALLOW_SYSCALL(recvmsg),
       //ALLOW_SYSCALL(shutdown),
       //ALLOW_SYSCALL(bind),
       //ALLOW_SYSCALL(listen),
       //ALLOW_SYSCALL(getsockname),
       //ALLOW_SYSCALL(getpeername),
       //ALLOW_SYSCALL(socketpair),
       //ALLOW_SYSCALL(setsockopt),
       //ALLOW_SYSCALL(getsockopt),
       ALLOW_SYSCALL(clone), // java
       //ALLOW_SYSCALL(fork),
       //ALLOW_SYSCALL(vfork),
       ALLOW_SYSCALL(execve), // general
       ALLOW_SYSCALL(exit), // general
       ALLOW_SYSCALL(wait4), // java
       //ALLOW_SYSCALL(kill),
       ALLOW_SYSCALL(uname), // java
       //ALLOW_SYSCALL(semget),
       //ALLOW_SYSCALL(semop),
       //ALLOW_SYSCALL(semctl),
       //ALLOW_SYSCALL(shmdt),
       //ALLOW_SYSCALL(msgget),
       //ALLOW_SYSCALL(msgsnd),
       //ALLOW_SYSCALL(msgrcv),
       //ALLOW_SYSCALL(msgctl),
       ALLOW_SYSCALL(fcntl), // general
       //ALLOW_SYSCALL(flock),
       //ALLOW_SYSCALL(fsync),
       //ALLOW_SYSCALL(fdatasync),
       //ALLOW_SYSCALL(truncate),
       //ALLOW_SYSCALL(ftruncate),
       ALLOW_SYSCALL(getdents), // general
       //ALLOW_SYSCALL(getcwd),
       //ALLOW_SYSCALL(chdir),
       //ALLOW_SYSCALL(fchdir),
       //ALLOW_SYSCALL(rename),
       //ALLOW_SYSCALL(mkdir),
       //ALLOW_SYSCALL(rmdir),
       //ALLOW_SYSCALL(creat),
       //ALLOW_SYSCALL(link),
       //ALLOW_SYSCALL(unlink),
       //ALLOW_SYSCALL(symlink),
       ALLOW_SYSCALL(readlink), // general
       //ALLOW_SYSCALL(chmod),
       //ALLOW_SYSCALL(fchmod),
       //ALLOW_SYSCALL(chown),
       //ALLOW_SYSCALL(fchown),
       //ALLOW_SYSCALL(lchown),
       //ALLOW_SYSCALL(umask),
       ALLOW_SYSCALL(gettimeofday), // java
       ALLOW_SYSCALL(getrlimit), // python
       ALLOW_SYSCALL(getrusage), // ruby
       //ALLOW_SYSCALL(sysinfo),
       //ALLOW_SYSCALL(times),
       //ALLOW_SYSCALL(ptrace),
       ALLOW_SYSCALL(getuid), // python
       //ALLOW_SYSCALL(syslog),
       ALLOW_SYSCALL(getgid), // python
       //ALLOW_SYSCALL(setuid),
       //ALLOW_SYSCALL(setgid),
       ALLOW_SYSCALL(geteuid), // python
       ALLOW_SYSCALL(getegid), // python
       //ALLOW_SYSCALL(setpgid),
       ALLOW_SYSCALL(getppid), // java
       ALLOW_SYSCALL(getpgrp), // java
       //ALLOW_SYSCALL(setsid),
       //ALLOW_SYSCALL(setreuid),
       //ALLOW_SYSCALL(setregid),
       //ALLOW_SYSCALL(getgroups),
       //ALLOW_SYSCALL(setgroups),
       //ALLOW_SYSCALL(setresuid),
       //ALLOW_SYSCALL(getresuid),
       //ALLOW_SYSCALL(setresgid),
       //ALLOW_SYSCALL(getresgid),
       //ALLOW_SYSCALL(getpgid),
       //ALLOW_SYSCALL(setfsuid),
       //ALLOW_SYSCALL(setfsgid),
       //ALLOW_SYSCALL(getsid),
       //ALLOW_SYSCALL(capget),
       //ALLOW_SYSCALL(capset),
       //ALLOW_SYSCALL(rt_sigpending),
       //ALLOW_SYSCALL(rt_sigtimedwait),
       //ALLOW_SYSCALL(rt_sigqueueinfo),
       //ALLOW_SYSCALL(rt_sigsuspend),
       ALLOW_SYSCALL(sigaltstack), // ruby
       //ALLOW_SYSCALL(utime),
       //ALLOW_SYSCALL(mknod),
       //ALLOW_SYSCALL(uselib),
       //ALLOW_SYSCALL(personality),
       //ALLOW_SYSCALL(ustat),
       //ALLOW_SYSCALL(statfs),
       //ALLOW_SYSCALL(fstatfs),
       //ALLOW_SYSCALL(sysfs),
       //ALLOW_SYSCALL(getpriority),
       //ALLOW_SYSCALL(setpriority),
       //ALLOW_SYSCALL(sched_setparam),
       //ALLOW_SYSCALL(sched_getparam),
       //ALLOW_SYSCALL(sched_setscheduler),
       //ALLOW_SYSCALL(sched_getscheduler),
       //ALLOW_SYSCALL(sched_get_priority_max),
       //ALLOW_SYSCALL(sched_get_priority_min),
       //ALLOW_SYSCALL(sched_rr_get_interval),
       //ALLOW_SYSCALL(mlock),
       //ALLOW_SYSCALL(munlock),
       //ALLOW_SYSCALL(mlockall),
       //ALLOW_SYSCALL(munlockall),
       //ALLOW_SYSCALL(vhangup),
       //ALLOW_SYSCALL(modify_ldt),
       //ALLOW_SYSCALL(pivot_root),
       //ALLOW_SYSCALL(_sysctl),
       //ALLOW_SYSCALL(prctl),
       ALLOW_SYSCALL(arch_prctl), // python
       //ALLOW_SYSCALL(adjtimex),
       //ALLOW_SYSCALL(setrlimit),
       //ALLOW_SYSCALL(chroot),
       //ALLOW_SYSCALL(sync),
       //ALLOW_SYSCALL(acct),
       //ALLOW_SYSCALL(settimeofday),
       //ALLOW_SYSCALL(mount),
       //ALLOW_SYSCALL(umount2),
       //ALLOW_SYSCALL(swapon),
       //ALLOW_SYSCALL(swapoff),
       //ALLOW_SYSCALL(reboot),
       //ALLOW_SYSCALL(sethostname),
       //ALLOW_SYSCALL(setdomainname),
       //ALLOW_SYSCALL(iopl),
       //ALLOW_SYSCALL(ioperm),
       //ALLOW_SYSCALL(create_module),
       //ALLOW_SYSCALL(init_module),
       //ALLOW_SYSCALL(delete_module),
       //ALLOW_SYSCALL(get_kernel_syms),
       //ALLOW_SYSCALL(query_module),
       //ALLOW_SYSCALL(quotactl),
       //ALLOW_SYSCALL(nfsservctl),
       //ALLOW_SYSCALL(getpmsg),
       //ALLOW_SYSCALL(putpmsg),
       //ALLOW_SYSCALL(afs_syscall),
       //ALLOW_SYSCALL(tuxcall),
       //ALLOW_SYSCALL(security),
       //ALLOW_SYSCALL(gettid),
       //ALLOW_SYSCALL(readahead),
       //ALLOW_SYSCALL(setxattr),
       //ALLOW_SYSCALL(lsetxattr),
       //ALLOW_SYSCALL(fsetxattr),
       //ALLOW_SYSCALL(getxattr),
       //ALLOW_SYSCALL(lgetxattr),
       //ALLOW_SYSCALL(fgetxattr),
       //ALLOW_SYSCALL(listxattr),
       //ALLOW_SYSCALL(llistxattr),
       //ALLOW_SYSCALL(flistxattr),
       //ALLOW_SYSCALL(removexattr),
       //ALLOW_SYSCALL(lremovexattr),
       //ALLOW_SYSCALL(fremovexattr),
       //ALLOW_SYSCALL(tkill),
       //ALLOW_SYSCALL(time),
       ALLOW_SYSCALL(futex), // general
       //ALLOW_SYSCALL(sched_setaffinity),
       ALLOW_SYSCALL(sched_getaffinity), // ruby
       //ALLOW_SYSCALL(set_thread_area),
       //ALLOW_SYSCALL(io_setup),
       //ALLOW_SYSCALL(io_destroy),
       //ALLOW_SYSCALL(io_getevents),
       //ALLOW_SYSCALL(io_submit),
       //ALLOW_SYSCALL(io_cancel),
       //ALLOW_SYSCALL(get_thread_area),
       //ALLOW_SYSCALL(lookup_dcookie),
       //ALLOW_SYSCALL(epoll_create),
       //ALLOW_SYSCALL(epoll_ctl_old),
       //ALLOW_SYSCALL(epoll_wait_old),
       //ALLOW_SYSCALL(remap_file_pages),
       //ALLOW_SYSCALL(getdents64),
       ALLOW_SYSCALL(set_tid_address), // python
       //ALLOW_SYSCALL(restart_syscall),
       //ALLOW_SYSCALL(semtimedop),
       //ALLOW_SYSCALL(fadvise64),
       //ALLOW_SYSCALL(timer_create),
       //ALLOW_SYSCALL(timer_settime),
       //ALLOW_SYSCALL(timer_gettime),
       //ALLOW_SYSCALL(timer_getoverrun),
       //ALLOW_SYSCALL(timer_delete),
       //ALLOW_SYSCALL(clock_settime),
       //ALLOW_SYSCALL(clock_gettime),
       //ALLOW_SYSCALL(clock_getres),
       ALLOW_SYSCALL(clock_nanosleep), // general
       ALLOW_SYSCALL(exit_group), // python
       //ALLOW_SYSCALL(epoll_wait),
       //ALLOW_SYSCALL(epoll_ctl),
       //ALLOW_SYSCALL(tgkill),
       //ALLOW_SYSCALL(utimes),
       //ALLOW_SYSCALL(vserver),
       //ALLOW_SYSCALL(mbind),
       //ALLOW_SYSCALL(set_mempolicy),
       //ALLOW_SYSCALL(get_mempolicy),
       //ALLOW_SYSCALL(mq_open),
       //ALLOW_SYSCALL(mq_unlink),
       //ALLOW_SYSCALL(mq_timedsend),
       //ALLOW_SYSCALL(mq_timedreceive),
       //ALLOW_SYSCALL(mq_notify),
       //ALLOW_SYSCALL(mq_getsetattr),
       //ALLOW_SYSCALL(kexec_load),
       //ALLOW_SYSCALL(waitid),
       //ALLOW_SYSCALL(add_key),
       //ALLOW_SYSCALL(request_key),
       //ALLOW_SYSCALL(keyctl),
       //ALLOW_SYSCALL(ioprio_set),
       //ALLOW_SYSCALL(ioprio_get),
       //ALLOW_SYSCALL(inotify_init),
       //ALLOW_SYSCALL(inotify_add_watch),
       //ALLOW_SYSCALL(inotify_rm_watch),
       //ALLOW_SYSCALL(migrate_pages),
       ALLOW_SYSCALL(openat), // python
       //ALLOW_SYSCALL(mkdirat),
       //ALLOW_SYSCALL(mknodat),
       //ALLOW_SYSCALL(fchownat),
       //ALLOW_SYSCALL(futimesat),
       //ALLOW_SYSCALL(newfstatat),
       //ALLOW_SYSCALL(unlinkat),
       //ALLOW_SYSCALL(renameat),
       //ALLOW_SYSCALL(linkat),
       //ALLOW_SYSCALL(symlinkat),
       //ALLOW_SYSCALL(readlinkat),
       //ALLOW_SYSCALL(fchmodat),
       //ALLOW_SYSCALL(faccessat),
       //ALLOW_SYSCALL(pselect6),
       //ALLOW_SYSCALL(ppoll),
       //ALLOW_SYSCALL(unshare),
       ALLOW_SYSCALL(set_robust_list), // python
       ALLOW_SYSCALL(get_robust_list),
       //ALLOW_SYSCALL(splice),
       //ALLOW_SYSCALL(tee),
       //ALLOW_SYSCALL(sync_file_range),
       //ALLOW_SYSCALL(vmsplice),
       //ALLOW_SYSCALL(move_pages),
       //ALLOW_SYSCALL(utimensat),
       //ALLOW_SYSCALL(epoll_pwait),
       //ALLOW_SYSCALL(signalfd),
       //ALLOW_SYSCALL(timerfd_create),
       //ALLOW_SYSCALL(eventfd),
       //ALLOW_SYSCALL(fallocate),
       //ALLOW_SYSCALL(timerfd_settime),
       //ALLOW_SYSCALL(timerfd_gettime),
       //ALLOW_SYSCALL(accept4),
       //ALLOW_SYSCALL(signalfd4),
       //ALLOW_SYSCALL(eventfd2),
       //ALLOW_SYSCALL(epoll_create1),
       //ALLOW_SYSCALL(dup3),
       //ALLOW_SYSCALL(pipe2),
       //ALLOW_SYSCALL(inotify_init1),
       //ALLOW_SYSCALL(preadv),
       //ALLOW_SYSCALL(pwritev),
       //ALLOW_SYSCALL(rt_tgsigqueueinfo),
       //ALLOW_SYSCALL(perf_event_open),
       //ALLOW_SYSCALL(recvmmsg),
       //ALLOW_SYSCALL(fanotify_init),
       //ALLOW_SYSCALL(fanotify_mark),
       //ALLOW_SYSCALL(prlimit64),
       //ALLOW_SYSCALL(name_to_handle_at),
       //ALLOW_SYSCALL(open_by_handle_at),
       //ALLOW_SYSCALL(clock_adjtime),
       //ALLOW_SYSCALL(syncfs),
       //ALLOW_SYSCALL(sendmmsg),
       //ALLOW_SYSCALL(setns),
       //ALLOW_SYSCALL(getcpu),
       //ALLOW_SYSCALL(process_vm_readv),
       //ALLOW_SYSCALL(process_vm_writev),
       //ALLOW_SYSCALL(kcmp),
       //ALLOW_SYSCALL(finit_module),
       //ALLOW_SYSCALL(sched_setattr),
       //ALLOW_SYSCALL(sched_getattr),
       //ALLOW_SYSCALL(renameat2),
       //ALLOW_SYSCALL(seccomp),
       KILL_PROCESS,
   };
   struct sock_fprog prog = {
       .len = (unsigned short)(sizeof(filter)/sizeof(filter[0])),
       .filter = filter,
   };

   if (prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)) {
       e("prctl(NO_NEW_PRIVS)");
       goto failed;
   }
   if (prctl(PR_SET_SECCOMP, SECCOMP_MODE_FILTER, &prog)) {
       e("prctl(SECCOMP)");
       goto failed;
   }
   return 0;

failed:
   if (errno == EINVAL)
       std::cerr << "SECCOMP_FILTER is not available. :(" << std::endl;
   return 1;
}

void limit_resource(int const resource, rlim_t const soft, rlim_t const hard) {
    rlimit const rl = { soft, hard };
    if (setrlimit(resource, &rl) != 0)
        e("setrlimit");
}

void limit_resource(int const resource, rlim_t const l) {
    limit_resource(resource, l, l);
}

void close_fds(void) {
    close(fileno(stdin));
    for (int fd = fileno(stderr); fd != 1024; ++fd)
        close(fd);
    dup2(fileno(stdout), fileno(stderr));
}

void install_rlimits(void) {
    limit_resource(RLIMIT_CPU, 10);                // 10 seconds CPU
    limit_resource(RLIMIT_AS, 400*1024*1024);      // 400MB address space
    limit_resource(RLIMIT_DATA, 400*1024*1024);    // 400MB data space
    limit_resource(RLIMIT_FSIZE, 10*1024*1024);    // Maximum filesize
    limit_resource(RLIMIT_LOCKS, 0);               // Maximum file locks held
    limit_resource(RLIMIT_MEMLOCK, 0);             // Maximum locked-in-memory address spac
    limit_resource(RLIMIT_NPROC, 0);               // Maximum number of processes.
}

int main(int const argc, char* const* const argv) {
    try {
        if (argc < 2)
            e("params: program args");

        close_fds();
        install_rlimits();
        if (install_syscall_filter())
            e("install_syscall_filter()");

        int rc = execv(argv[1], &(argv[1]));
        if (rc == -1)
            e(std::strerror(errno));

    } catch (std::exception const & e) {
        std::cerr << "exception: " << e.what() << std::endl;
        return 1;
    }
}