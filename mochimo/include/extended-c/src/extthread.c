/**
 * @private
 * @headerfile extthread.h <extthread.h>
 * @copyright Adequate Systems LLC, 2018-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_THREAD_C
#define EXTENDED_THREAD_C


#include "extthread.h"

/**
 * Initialize a Mutex.
 * @param mutexp Pointer to a ::Mutex
 * @returns 0 on success, else error number.
 * @note On Windows, this function uses a static spinlock to
 * guard the initialization of all mutexes, either through direct
 * use of this function or via the first call to mutex_lock().
 * Although negligible, it should be noted that this causes a
 * synchronous effect for all such calls.
*/
int mutex_init(Mutex *mutexp)
{
#if OS_WINDOWS
   /* CRITICAL_SECTION initialization is guarded by an initialization
    * spinlock, which is 32-bit aligned as required by the Windows
    * API's InterlockedCompareExchange() atomic function. */
   __declspec(align(4)) static volatile LONG spinlock = 0;

   /* per pthreads_mutex_init(), return EBUSY for initialized mutex */
   if (mutexp->init > 0) return EBUSY;
   /* acquire exclusive spin lock for initialisation */
   while(InterlockedCompareExchange(&spinlock, 1, 0));
   /* vv SPINLOCKED SECTION START */

   /* attempt initialization if uninitialised */
   if(mutexp->init < 1) {
      /* initialize critical section inside Mutex */
      InitializeCriticalSection(&(mutexp->lock));
      /* set Mutex initialized */
      mutexp->init = 1;
   }

   /* ^^ END SPINLOCK SECTION */
   /* release exclusive spin lock */
   spinlock = 0;

   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_mutex_init(mutexp, NULL);

#endif
}

/**
 * Acquire an exclusive lock on a Mutex.
 * @param mutexp Pointer to a ::Mutex
 * @returns 0 on success, else error number.
 * @note On Windows, a statically initialized Mutex will be
 * initialized on the first call to mutex_lock().
*/
int mutex_lock(Mutex *mutexp)
{
#if OS_WINDOWS
   /* initialize mutex if statically initialized */
   if (mutexp->init < 1) mutex_init(mutexp);
   /* acquire exclusive critical section lock */
   EnterCriticalSection(&(mutexp->lock));

   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_mutex_lock(mutexp);

#endif
}

/**
 * Release an exclusive lock on a Mutex.
 * @param mutexp Pointer to a ::Mutex
 * @returns 0 on success, else error number.
*/
int mutex_unlock(Mutex *mutexp)
{
#if OS_WINDOWS
   LeaveCriticalSection(&(mutexp->lock));
   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_mutex_unlock(mutexp);

#endif
}

/**
 * Destroy a Mutex. Free's any memory allocated by the system,
 * and marks Mutex as uninitialized.
 * @param mutexp Pointer to a ::Mutex
 * @returns 0 on success, else error number.
*/
int mutex_destroy(Mutex *mutexp)
{
#if OS_WINDOWS
   /* destroy critical section */
   DeleteCriticalSection(&(mutexp->lock));
   /* set Mutex uninitialized */
   mutexp->init = 0;

   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_mutex_unlock(mutexp);

#endif
}

/**
 * Initialize a Read Write Lock.
 * @param rwlockp Pointer to a ::RWLock
 * @returns 0 on success, else error code.
*/
int rwlock_init(RWLock *rwlockp)
{
#if OS_WINDOWS
   InitializeSRWLock(rwlockp);
   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_rwlock_init(rwlockp, NULL);

#endif
}

/**
 * Lock a ::RWLock object for reading. The calling thread acquires the
 * read lock if no write locks are held or waiting on the lock.
 * @param rwlockp Pointer to a ::RWLock
 * @returns 0 on success, else error code.
 * @note This function can block the execution of a thread.
*/
int rwlock_rdlock(RWLock *rwlockp)
{
#if OS_WINDOWS
   AcquireSRWLockShared(rwlockp);
   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_rwlock_rdlock(rwlockp);

#endif
}

/**
 * Lock a ::RWLock object for writing. The calling thread acquires the
 * write lock if no read or write locks are held or waiting on the lock.
 * @param rwlockp Pointer to a ::RWLock
 * @returns 0 on success, else error code.
 * @note This function can block the execution of a thread.
*/
int rwlock_wrlock(RWLock *rwlockp)
{
#if OS_WINDOWS
   AcquireSRWLockExclusive(rwlockp);
   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_rwlock_wrlock(rwlockp);

#endif
}

/**
 * Release a ::RWLock object locked for reading.
 * @param rwlockp Pointer to a ::RWLock
 * @returns 0 on success, else error code.
*/
int rwlock_rdunlock(RWLock *rwlockp)
{
#if OS_WINDOWS
   ReleaseSRWLockShared(rwlockp);
   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_rwlock_unlock(rwlockp);

#endif
}

/**
 * Release a ::RWLock object locked for writing.
 * @param rwlockp Pointer to a ::RWLock
 * @returns 0 on success, else error code.
*/
int rwlock_wrunlock(RWLock *rwlockp)
{
#if OS_WINDOWS
   ReleaseSRWLockExclusive(rwlockp);
   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_rwlock_unlock(rwlockp);

#endif
}

/**
 * Destroy a Read Write Lock. Free's any memory allocated by the system.
 * @param rwlockp Pointer to a ::RWLock
 * @returns 0 on success, else error number.
*/
int rwlock_destroy(RWLock *rwlockp)
{
#if OS_WINDOWS
   /* A SRWLock does not need to be explicitly destroyed, so this
    * function simply tries to acquire an exclusive lock on *rwlock
    * and, either immediately releases the lock, or return the
    * EBUSY code, immitating pthread_rwlock_destory() operation. */
   if (TryAcquireSRWLockExclusive(rwlockp)) {
      ReleaseSRWLockExclusive(rwlockp);
      return 0;
   }

   return EBUSY;

#elif defined(_POSIX_THREADS)
   return pthread_rwlock_destroy(rwlockp);

#endif
}

/**
 * Create/Start a new thread. Thread identifier is placed in @a tidp.
 * @param tidp Pointer to ThreadId to place thread identifier.
 * @param fnp ::ThreadProc function to execute on new thread.
 * @param argp Pointer to arguments to give to @a fnp.
 * @returns 0 on success, else error number
*/
int thread_create(ThreadId *tidp, ThreadRoutine fnp, void *argp)
{
#if OS_WINDOWS
   HANDLE thandle = CreateThread(NULL, 0, fnp, argp, 0, tidp);
   if (thandle == NULL) return GetLastError();
   CloseHandle(thandle);

   return 0;

#elif defined(_POSIX_THREADS)
   return pthread_create(tidp, NULL, fnp, argp);

#endif
}  /* end thread_create() */

/**
 * Join with a terminated thread.
 * Waits for the thread specified by @a tid to terminate.
 * @param tid A ::ThreadId
 * @returns 0 on success, else error number.
*/
int thread_join(ThreadId tid)
{
#if OS_WINDOWS
   HANDLE thandle;
   int ecode = 0;

   /* acquire Thread HANDLE from thread id */
   thandle = OpenThread(SYNCHRONIZE, FALSE, tid);
   if(thandle == NULL) return GetLastError();
   /* wait indefinitely for thread to complete */
   if(WaitForSingleObject(thandle, INFINITE) == WAIT_FAILED) {
      ecode = GetLastError();
   }
   /* close handle to thread */
   CloseHandle(thandle);

#elif defined(_POSIX_THREADS)
   return pthread_join(tid, NULL);

#endif
}  /* end thread_join() */

/**
 * Join with a list of terminated threads.
 * Waits for all threads in a list specified by tidlist[count] to
 * terminate and "joins" with them in a sequential manner.
 * @param tidlist Pointer to a list of ::ThreadId's
 * @param count Number of threads in list
 * @returns 0 on success, else the first error which occurred.
 * @note If a ::ThreadId within the @a tidlist evaluates as Zero (0),
 * that ::ThreadId will be skipped.
*/
int thread_join_list(ThreadId tidlist[], int count)
{
   int i, temp, ecode;

   for(ecode = temp = i = 0; i < count; i++) {
      if (tidlist[i]) temp = thread_join(tidlist[i]);
      if (temp && !ecode) ecode = temp;
   }

   return ecode;
}

/**
 * Send a termination request to a thread. Does not wait for termination.
 * @param tid A ::ThreadId
 * @returns 0 on success, else error number.
 * @note It is not recommended to forcefully terminate threads in running
 * processes, and should instead be reserved for (ungraceful) shutdowns.
*/
int thread_terminate(ThreadId tid)
{
#if OS_WINDOWS
   HANDLE thandle;
   int ecode;

   /* acquire Thread HANDLE from thread id */
   thandle = OpenThread(SYNCHRONIZE, FALSE, tid);
   if(thandle == NULL) return GetLastError();
   /* wait indefinitely for thread to complete */
   ecode = TerminateThread(thandle, 0) ? 0 : GetLastError();
   /* close handle to thread */
   CloseHandle(thandle);

#elif defined(_POSIX_THREADS)
   return pthread_cancel(tid);

#endif
}  /* end thread_terminate() */

/**
 * Terminate a list of threads. Sends a termination request to all threads
 * in a list specified by @a tidlist[count]. Does not wait for termination.
 * @param tidlist Pointer to a list of ::ThreadId's
 * @param count Number of threads in list
 * @returns 0 on success, else the first error which occurred.
 * @note If a ::ThreadId within the @a tidlist evaluates as Zero (0),
 * that ::ThreadId will be skipped.
*/
int thread_terminate_list(ThreadId tidlist[], int count)
{
   int i, temp, ecode;

   for(ecode = temp = i = 0; i < count; i++) {
      if (tidlist[i]) temp = thread_terminate(tidlist[i]);
      if (temp && !ecode) ecode = temp;
   }

   return ecode;
}  /* end thread_terminate_list() */

/* end include guard */
#endif
