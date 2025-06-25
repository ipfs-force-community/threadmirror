import { useState } from "react";
import { useAuth0 } from "@auth0/auth0-react";
import { UserIcon } from "./icons";
import styles from "./UserLoginOut.module.css";



const UserLgoinOut = () => {
    const { user, isAuthenticated, loginWithRedirect, logout, getAccessTokenSilently } = useAuth0();
    const [accessToken, setAccessToken] = useState<string | null>(null);
    const logoutWithRedirect = () =>
        logout({
            logoutParams: {
                returnTo: window.location.origin,
            },
        });

    console.log("user rendered", { user, isAuthenticated });
    if (isAuthenticated) {
        if (!accessToken) {
            getAccessTokenSilently().then(token => {
                console.info("Access Token:", JSON.stringify(token));
                setAccessToken(token);
            }).catch(err => {
                console.error("Error fetching access token:", err);
            });
        }
    }

    return (
        <div className={styles.nav_container}>
            <div
                role="button"
                title={isAuthenticated ? "Logout" : "Login"}
                onClick={() => isAuthenticated ? logoutWithRedirect() : loginWithRedirect()}
                className={styles.user_button}
            >
                <UserIcon isLoggedIn={isAuthenticated} />
            </div>
        </div>
    );
};

export default UserLgoinOut;