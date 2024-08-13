/**
 * @file extinet.h
 * @brief Extended internet support.
 * @details Provides extended support for internet related functionality.
 * @copyright Adequate Systems LLC, 2018-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
 * @note Applications MUST call sock_startup() before using some
 * socket support functions in this file.
*/

/* include guard */
#ifndef EXTENDED_INTERNET_H
#define EXTENDED_INTERNET_H


#include "extos.h"

#if OS_WINDOWS
   #pragma comment(lib, "ws2_32.lib")
   #include <winsock2.h>  /* for socket handling */
   #include <ws2tcpip.h>  /* for socklen_t */

   #define os_dependant_sock_errno  ( WSAGetLastError() )
   #define sock_err_is_success(_e)  ( _e == WSAEISCONN )
   #define sock_err_is_waiting(_e) \
      ( _e == WSAEWOULDBLOCK || _e == WSAEALREADY || _e == WSAEINVAL )

/* end OS_WINDOWS */
#elif OS_UNIX
   #include <arpa/inet.h>
   #include <errno.h>
   #include <netdb.h>
   #include <netinet/in.h>
   #include <sys/socket.h>
   #include <unistd.h>

   #define os_dependant_sock_errno  ( errno )
   #define sock_err_is_success(_e)  ( _e == EISCONN )
   #define sock_err_is_waiting(_e)  ( _e == EINPROGRESS || _e == EALREADY )

/* end Unix */
#endif

#define WEBDOMAIN_MAX   256    /**< Website domain name byte length limit. */
#define WEBPATH_MAX     8196   /**< Website path byte length limit. */

/**
 * Detailed socket error code.
 * <br/>On Windows, expands to: `( WSAGetLastError() )`
 * <br/>On Unix, expands to: `( errno )`
*/
#define sock_errno   os_dependant_sock_errno

#ifndef SOCKET
/**
 * SOCKET datatype for use with socket support functions
 * @note SOCKET is ONLY guaranteed an `int` type where NOT already defined.
*/
#define SOCKET int
#endif

#ifndef SOCKET_ERROR
#define SOCKET_ERROR (-1)  /**< Socket error return code */
#endif

#ifndef INVALID_SOCKET
#define INVALID_SOCKET (SOCKET)(~0) /**< Indicates empty socket descriptor */
#endif

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

int http_get(char *url, char *fname, int timeout);
char *ntoa(void *n, char *a);
unsigned long aton(char *a);
unsigned long get_sock_ip(SOCKET sd);
int sock_startup(void);
int sock_state(void);
int sock_cleanup(void);
int sock_set_nonblock(SOCKET sd);
int sock_set_blocking(SOCKET sd);
int sock_close(SOCKET sd);
SOCKET sock_connect_ip(unsigned long ip, unsigned short port, double timeout);
SOCKET sock_connect_addr(char *addr, unsigned short port, double timeout);
int sock_recv(SOCKET sd, void *pkt, int len, int flags, double timeout);
int sock_send(SOCKET sd, void *pkt, int len, int flags, double timeout);
int gethostip(char *name, int namelen);

#ifdef __cplusplus
}  /* end extern "C" */
#endif

/* end include guard */
#endif
