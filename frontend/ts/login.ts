// Handle login/register dynamics
(() => {
    let status: "register" | "login" =
        window.location.pathname === "/user/register" ? "register" : "login";

    const loginForm = document.getElementById("login-form") as HTMLFormElement;
    const loginState = document.getElementById("login-state") as HTMLDivElement;
    const loginSwitch = document.getElementById(
        "login-switch",
    ) as HTMLButtonElement;
    const registerFields = document.getElementById(
        "register-fields",
    ) as HTMLDivElement;

    const apply = () => {
        if (status === "register") {
            loginForm.action = "/user/register";
            loginState.innerHTML = "Register";
            loginSwitch.innerHTML = "Logging in?";
            registerFields.classList.remove("hidden");
        } else {
            loginForm.action = "/user/login";
            loginState.innerHTML = "Login";
            loginSwitch.innerHTML = "New user?";
            registerFields.classList.add("hidden");
        }
        loginForm.action += window.location.search;
    };

    apply();

    loginSwitch.style.display = "";
    loginSwitch.addEventListener("click", () => {
        status = status === "login" ? "register" : "login";
        apply();
    });
})();
