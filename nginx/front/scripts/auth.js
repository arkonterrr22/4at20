function auth_check() {
    const token = localStorage.getItem("token");
    const path = window.location.pathname;
    console.log("---\npath: ", path, "\ntoken: ", token, "\n---\n\n")
    if (!token) {
        console.log("!no token")
        // localStorage.removeItem("token");
        if (!path.endsWith("/login.html") && !path.endsWith("/register.html")) {
            console.log("!redirecting to /login")
            window.location.href = "./login.html";
        }
        return;
    }

    else {
            const decodedtoken = jwtDecode(token);
            if (path.endsWith("/login.html") || path.endsWith("/register.html")) {
            console.log("!redirecting to /")
            window.location.href = "/";
        }
        window.token = decodedtoken
        return; 
    }       
}

auth_check();
