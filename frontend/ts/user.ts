// Change Password validation
(() => {
    let form = document.getElementById(
        "change-password-form",
    ) as HTMLFormElement | null;
    if (!form) return;

    let newPassword = document.getElementById(
        "new-password",
    ) as HTMLInputElement;
    let confirmPassword = document.getElementById(
        "confirm-password",
    ) as HTMLInputElement;
    let matchWarning = document.getElementById(
        "match_warning",
    ) as HTMLSpanElement;

    form.addEventListener("submit", (e) => {
        if (newPassword.value !== confirmPassword.value) {
            e.preventDefault();
            matchWarning.style.display = "inline";
            confirmPassword.classList.add("border", "border-red-600");
        }
    });
})();
