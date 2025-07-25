import { useState, useEffect, useCallback, useRef } from "react";
import { useAuth0, User } from "@auth0/auth0-react";
import { UserIcon } from "./icons";
import { toast } from 'sonner';
import {
    getUserInfo,
    saveAuthCookies,
    clearAuthCookies,
    isUserLoggedIn
} from '@utils/cookie';
import styles from "./UserLoginOut.module.css";
import defaultProfile from '../default_profile.png';
import { useAuthContext } from '../AuthContext';

const UserLgoinOut = () => {
    const { user, error, getAccessTokenSilently, loginWithRedirect, logout, isAuthenticated, isLoading: auth0Loading } = useAuth0();
    const { setIsLoggedIn } = useAuthContext();
    const [localUser, setLocalUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const lastErrorRef = useRef<string | null>(null);

    const restoreUserFromCookies = useCallback(() => {
        if (!isUserLoggedIn()) return false;

        const storedUser = getUserInfo<User>();
        if (storedUser) {
            setLocalUser(storedUser);
            return true;
        }

        return false;
    }, []);

    const saveAuthData = useCallback(async () => {
        if (user) {
            try {
                const token = await getAccessTokenSilently();
                const userToStore = {
                    sub: user.sub,
                    name: user.name,
                    email: user.email,
                    picture: user.picture,
                    nickname: user.nickname,
                    updated_at: user.updated_at
                };

                saveAuthCookies(token, userToStore);
                setLocalUser(user);
                setIsLoggedIn(true); // 登录后同步 context
                return true;
            } catch (error) {
                console.error("Authentication data save failed", error);
            }
        }
        return false;
    }, [user, getAccessTokenSilently, setIsLoggedIn]);

    useEffect(() => {
        const initAuth = async () => {
            setIsLoading(true);

            const restored = restoreUserFromCookies();

            if (!restored && user) {
                await saveAuthData();
            } else if (restored) {
                setIsLoggedIn(true); // 恢复登录状态
            } else {
                setIsLoggedIn(false);
            }

            setIsLoading(false);
        };

        initAuth();
    }, [user, restoreUserFromCookies, saveAuthData, setIsLoggedIn]);

    useEffect(() => {
        if (error && error.message) {
            if (lastErrorRef.current !== error.message) {
                toast.error(`Authentication error: ${error.message}`, {
                    duration: 5000,
                    id: `auth-error-${Date.now()}`
                });

                lastErrorRef.current = error.message;

                setTimeout(() => {
                    if (lastErrorRef.current === error.message) {
                        lastErrorRef.current = null;
                    }
                }, 5000);
            }
        }
    }, [error]);

    const handleLogout = () => {
        clearAuthCookies();
        setLocalUser(null);
        setIsLoggedIn(false); // 登出时同步 context
        logout({ logoutParams: { returnTo: window.location.origin } });
    };

    const handleLogin = () => {
        loginWithRedirect().catch((error) => {
            console.error("Login error:", error);
            toast.error(`Login failed: ${error.message}`);
        });
    };

    const handleAuthAction = () => {
        if (isAuthenticated && isUserLoggedIn()) {
            handleLogout();
        } else {
            handleLogin();
        }
    };

    const handleLogoutClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        handleLogout();
    };

    if (isLoading || auth0Loading) {
        return <div className="flex justify-center p-4">Loading...</div>;
    }

    const displayUser = localUser || user;
    const isFullyLoggedIn = isAuthenticated && isUserLoggedIn();

    return (
        <div className={styles.nav_container}>
            <div className={styles.user_container}>
                {isFullyLoggedIn && displayUser && (
                    <div className={styles.user_profile} >
                        {displayUser.picture ? (
                            <img
                                src={displayUser.picture}
                                alt="User Avatar"
                                className={styles.user_avatar}
                            />
                        ) : (
                            <img
                                src={defaultProfile}
                                alt="User Avatar"
                                className={styles.user_avatar}
                            />
                        )}

                        <div className={styles.user_info_container}>
                            <span className={styles.user_name_text}>{displayUser.name}</span>
                            <span className={styles.user_nickname}>
                                <a href={`https://x.com/${displayUser.nickname}`} target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none', color: '#1da1f2' }}>
                                    @{displayUser.nickname}
                                </a>
                            </span>
                        </div>

                        <span className={styles.logout_icon} onClick={handleLogoutClick} title="Logout">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path d="M9 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                                <path d="M16 17L21 12L16 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                                <path d="M21 12H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                            </svg>
                        </span>
                    </div>
                )}
                {!isFullyLoggedIn && (
                    <div
                        role="button"
                        title="Login"
                        onClick={handleAuthAction}
                        className={styles.user_button}
                    >
                        <UserIcon isLoggedIn={false} />
                    </div>
                )}
            </div>
        </div>
    );
};

export default UserLgoinOut;