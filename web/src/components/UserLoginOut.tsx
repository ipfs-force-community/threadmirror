import { useState, useEffect, useCallback, useRef } from "react";
import { useAuth0, User } from "@auth0/auth0-react";
import { UserIcon } from "./icons";
import { useCookies } from 'react-cookie';
import { toast } from 'sonner';
import styles from "./UserLoginOut.module.css";

// 延长cookie过期时间到7天
const COOKIE_OPTIONS = {
    path: '/',
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax' as 'lax',
    maxAge: 7 * 24 * 60 * 60,
    domain: window.location.hostname
};

const UserLgoinOut = () => {
    const { isAuthenticated, user, error, getAccessTokenSilently, loginWithRedirect, logout } = useAuth0();
    const [cookies, setCookie, removeCookie] = useCookies(['auth_token', 'user_info']);
    const [localUser, setLocalUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [showUserInfo, setShowUserInfo] = useState(false);
    const lastErrorRef = useRef<string | null>(null);

    if (cookies.auth_token) {
        console.log('auth_token', cookies.auth_token);
    }

    const restoreUserFromCookies = useCallback(() => {
        if (!cookies.auth_token) return false;

        if (cookies.user_info) {
            try {
                const storedUser = typeof cookies.user_info === 'string'
                    ? JSON.parse(cookies.user_info)
                    : cookies.user_info;

                if (storedUser && storedUser.sub) {
                    setLocalUser(storedUser);
                    return true;
                }
            } catch (error) {
                removeCookie('auth_token', { path: '/' });
                removeCookie('user_info', { path: '/' });
            }
        }
        return false;
    }, [cookies, removeCookie]);

    const saveAuthData = useCallback(async () => {
        if (isAuthenticated && user) {
            try {
                const token = await getAccessTokenSilently();

                setCookie('auth_token', token, COOKIE_OPTIONS);

                const userToStore = {
                    sub: user.sub,
                    name: user.name,
                    email: user.email,
                    picture: user.picture,
                    nickname: user.nickname,
                    updated_at: user.updated_at
                };

                const userJson = JSON.stringify(userToStore);
                setCookie('user_info', userJson, COOKIE_OPTIONS);

                setLocalUser(user);
                return true;
            } catch (error) {
                console.error("Authentication data save failed", error);
            }
        }
        return false;
    }, [isAuthenticated, user, getAccessTokenSilently, setCookie]);

    useEffect(() => {
        const initAuth = async () => {
            setIsLoading(true);

            const restored = restoreUserFromCookies();

            if (!restored && isAuthenticated && user) {
                await saveAuthData();
            }

            setIsLoading(false);
        };

        initAuth();
    }, [isAuthenticated, user, restoreUserFromCookies, saveAuthData, cookies.auth_token]);

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
        removeCookie('auth_token', { path: '/' });
        removeCookie('user_info', { path: '/' });
        setLocalUser(null);
        setShowUserInfo(false);
        logout({ logoutParams: { returnTo: window.location.origin } });
    };

    const handleLogin = () => {
        loginWithRedirect().catch((error) => {
            console.error("Login error:", error);
            toast.error(`Login failed: ${error.message}`);
        });
    };

    const handleAuthAction = () => {
        if (isAuthenticated || localUser) {
            handleLogout();
        } else {
            handleLogin();
        }
    };

    const toggleUserInfo = (e: React.MouseEvent) => {
        e.stopPropagation();
        if (isAuthenticated || localUser) {
            setShowUserInfo(!showUserInfo);
        }
    };

    const handleLogoutClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        handleLogout();
    };

    if (isLoading) {
        return <div className="flex justify-center p-4">Loading...</div>;
    }

    const isLoggedIn = isAuthenticated || !!localUser;
    const displayUser = localUser || user;

    return (
        <div className={styles.nav_container}>
            <div className={styles.user_container}>
                {isLoggedIn && displayUser && (
                    <div className={styles.user_profile} onClick={toggleUserInfo}>
                        {displayUser.picture && (
                            <img
                                src={displayUser.picture}
                                alt="User Avatar"
                                className={styles.user_avatar}
                            />
                        )}
                        
                        <div className={styles.user_info_container}>
                            <span className={styles.user_name_text}>{displayUser.name}</span>
                            <span className={styles.user_nickname}>
                                <a href={`https://twitter.com/${displayUser.nickname}`} target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none', color: '#1da1f2' }}>
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
                {!isLoggedIn && (
                    <div
                        role="button"
                        title="Login"
                        onClick={handleAuthAction}
                        className={styles.user_button}
                    >
                        <UserIcon isLoggedIn={isLoggedIn} />
                    </div>
                )}
            </div>
        </div>
    );
};

export default UserLgoinOut;