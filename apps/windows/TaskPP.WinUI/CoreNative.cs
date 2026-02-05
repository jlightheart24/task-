using System;
using System.Runtime.InteropServices;

namespace TaskPP.WinUI;

internal static class CoreNative
{
    private const string DllName = "taskppcore.dll";

    [DllImport(DllName, EntryPoint = "Core_New", CallingConvention = CallingConvention.Cdecl)]
    public static extern ulong Core_New([MarshalAs(UnmanagedType.LPUTF8Str)] string configJson);

    [DllImport(DllName, EntryPoint = "Core_Destroy", CallingConvention = CallingConvention.Cdecl)]
    public static extern void Core_Destroy(ulong handle);

    [DllImport(DllName, EntryPoint = "Core_Open", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_Open(ulong handle);

    [DllImport(DllName, EntryPoint = "Core_Close", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_Close(ulong handle);

    [DllImport(DllName, EntryPoint = "Core_InitKeys", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_InitKeys(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string passphrase);

    [DllImport(DllName, EntryPoint = "Core_UnlockKeys", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_UnlockKeys(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string passphrase);

    [DllImport(DllName, EntryPoint = "Core_ListTasks", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_ListTasks(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string filterJson);

    [DllImport(DllName, EntryPoint = "Core_CreateTask", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_CreateTask(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string taskJson);

    [DllImport(DllName, EntryPoint = "Core_UpdateTask", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_UpdateTask(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string taskJson);

    [DllImport(DllName, EntryPoint = "Core_DeleteTask", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_DeleteTask(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string taskId);

    [DllImport(DllName, EntryPoint = "Core_ReorderTasks", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_ReorderTasks(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string reorderJson);

    [DllImport(DllName, EntryPoint = "Core_SetDueDate", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_SetDueDate(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string taskId, [MarshalAs(UnmanagedType.LPUTF8Str)] string dueDate);

    [DllImport(DllName, EntryPoint = "Core_SetCompleted", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_SetCompleted(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string taskId, int completed);

    [DllImport(DllName, EntryPoint = "Core_ExportEvents", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_ExportEvents(ulong handle, long sinceSeq);

    [DllImport(DllName, EntryPoint = "Core_ImportEvents", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_ImportEvents(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string eventsJson);

    [DllImport(DllName, EntryPoint = "Core_GetSyncState", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_GetSyncState(ulong handle);

    [DllImport(DllName, EntryPoint = "Core_DebugDecryptEvent", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr Core_DebugDecryptEvent(ulong handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string payloadBase64);

    [DllImport(DllName, EntryPoint = "Core_FreeString", CallingConvention = CallingConvention.Cdecl)]
    public static extern void Core_FreeString(IntPtr str);

    public static string ReadAndFree(IntPtr ptr)
    {
        if (ptr == IntPtr.Zero)
        {
            return string.Empty;
        }
        var value = Marshal.PtrToStringUTF8(ptr) ?? string.Empty;
        Core_FreeString(ptr);
        return value;
    }
}
