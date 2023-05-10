function Add-Path($Path) {
    $Path =
        [Environment]::GetEnvironmentVariable("PATH", "Machine") +
        $Path +
        [IO.Path]::PathSeparator
    [Environment]::SetEnvironmentVariable("Path", $Path, "Machine")
}
