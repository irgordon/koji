/* ============================================================
 * KOJI Kernel — AMD64 Linker Script
 *
 * Higher-half kernel at 0xFFFFFFFF80000000.
 * Bootloader (Limine/multiboot2) loads at physical 0x100000
 * and identity-maps + higher-half maps for us.
 * ============================================================ */

ENTRY(_start)

KERNEL_VMA = 0xFFFFFFFF80000000;
KERNEL_LMA = 0x100000;

SECTIONS
{
    . = KERNEL_VMA;

    .text ALIGN(4K) : AT(KERNEL_LMA)
    {
        __text_start = .;
        *(.text .text.*)
        __text_end = .;
    }

    .rodata ALIGN(4K) : AT(KERNEL_LMA + (__rodata_start - KERNEL_VMA))
    {
        __rodata_start = .;
        *(.rodata .rodata.*)
        __rodata_end = .;
    }

    .data ALIGN(4K) : AT(KERNEL_LMA + (__data_start - KERNEL_VMA))
    {
        __data_start = .;
        *(.data .data.*)
        __data_end = .;
    }

    .bss ALIGN(4K) : AT(KERNEL_LMA + (__bss_start - KERNEL_VMA))
    {
        __bss_start = .;
        *(.bss .bss.*)
        *(COMMON)
        __bss_end = .;
    }

    __kernel_end = .;

    /DISCARD/ :
    {
        *(.comment)
        *(.note*)
        *(.eh_frame*)
    }
}
