/* ----------------------------------------------------------------------------
 * This file was automatically generated by SWIG (http://www.swig.org).
 * Version 2.0.11
 *
 * This file is not intended to be easily readable and contains a number of
 * coding conventions designed to improve portability and efficiency. Do not make
 * changes to this file unless you know what you are doing--modify the SWIG
 * interface file instead.
 * ----------------------------------------------------------------------------- */

/* This file should be compiled with 6c/8c.  */
#pragma dynimport _ _ "sub.so"

#include "runtime.h"
#include "cgocall.h"

#ifdef _64BIT
#define SWIG_PARM_SIZE 8
#else
#define SWIG_PARM_SIZE 4
#endif

#pragma dynimport _wrap_new_sub _wrap_new_sub ""
extern void (*_wrap_new_sub)(void*);
static void (*x_wrap_new_sub)(void*) = _wrap_new_sub;

void
·_swig_wrap_new_sub(struct {
  uint8 x[SWIG_PARM_SIZE + SWIG_PARM_SIZE];
} p)

{
  runtime·cgocall(x_wrap_new_sub, &p);
}



#pragma dynimport _wrap_sub_Init _wrap_sub_Init ""
extern void (*_wrap_sub_Init)(void*);
static void (*x_wrap_sub_Init)(void*) = _wrap_sub_Init;

void
·_swig_wrap_sub_Init(struct {
  uint8 x[SWIG_PARM_SIZE + SWIG_PARM_SIZE + SWIG_PARM_SIZE + SWIG_PARM_SIZE];
} p)

{
  runtime·cgocall(x_wrap_sub_Init, &p);
}



#pragma dynimport _wrap_delete_sub _wrap_delete_sub ""
extern void (*_wrap_delete_sub)(void*);
static void (*x_wrap_delete_sub)(void*) = _wrap_delete_sub;

void
·_swig_wrap_delete_sub(struct {
  uint8 x[SWIG_PARM_SIZE + SWIG_PARM_SIZE];
} p)

{
  runtime·cgocall(x_wrap_delete_sub, &p);
}



