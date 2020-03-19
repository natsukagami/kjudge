interface Window {
    handleRegistration: () => void;
}

// Handle login/register dynamics
window.handleRegistration = () => {
    let status: "register" | "login" =
        window.location.pathname === "/user/register" ? "register" : "login";

    const loginForm = document.getElementById("login-form") as HTMLFormElement;
    const loginState = document.getElementById("login-state") as HTMLDivElement;
    const loginSwitch = document.getElementById(
        "login-switch",
    ) as HTMLButtonElement;

    const apply = () => {
        if (status === "register") {
            loginForm.action = "/user/register";
            loginState.innerHTML = "Register";
            loginSwitch.innerHTML = "Logging in?";
        } else {
            loginForm.action = "/user/login";
            loginState.innerHTML = "Login";
            loginSwitch.innerHTML = "New user?";
        }
        loginForm.action += window.location.search;
    };

    apply();

    loginSwitch.style.display = "";
    loginSwitch.addEventListener("click", () => {
        status = status === "login" ? "register" : "login";
        apply();
    });
};
