import React from 'react';
import { useAuth0 } from "@auth0/auth0-react";
import { toast } from 'sonner';
import styles from './LoginPrompt.module.css';

const LoginPrompt: React.FC = () => {
  const { loginWithRedirect } = useAuth0();

  const handleLogin = async () => {
    try {
      await loginWithRedirect({
        appState: {
          returnTo: window.location.pathname
        },
        authorizationParams: {
          prompt: 'login'
        }
      });
    } catch (error) {
      console.error("Login error:", error);
      toast.error(`Login failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.content}>
        <div className={styles.hero}>
          <h1 className={styles.title}>Decentralize your archives.</h1>
          <p className={styles.subtitle}>
            One-step archive to store and preserve content forever on the blockchain. Powered by <strong>Filecoin's</strong> decentralized storage network.
          </p>
        </div>
        
        <div className={styles.features}>
          <div className={styles.feature}>
            <div className={styles.featureIcon}>🌐</div>
            <h3 className={styles.featureTitle}><strong>Filecoin</strong> Storage</h3>
            <p className={styles.featureText}>
              Your content is stored on <strong>Filecoin's</strong> decentralized network, ensuring permanent availability and censorship resistance.
            </p>
          </div>
          
          <div className={styles.feature}>
            <div className={styles.featureIcon}>🔐</div>
            <h3 className={styles.featureTitle}>Blockchain Verified</h3>
            <p className={styles.featureText}>
              Every archive is cryptographically verified and immutably stored on the blockchain for ultimate trust.
            </p>
          </div>
          
          <div className={styles.feature}>
            <div className={styles.featureIcon}>⚡</div>
            <h3 className={styles.featureTitle}>One-Step Archive</h3>
            <p className={styles.featureText}>
              Single mention instantly creates permanent blockchain archives - no complex setup or multiple steps required.
            </p>
          </div>
        </div>

        <div className={styles.howItWorks}>
          <h2 className={styles.sectionTitle}>How blockchain mirroring works</h2>
          <div className={styles.steps}>
            <div className={styles.step}>
              <span className={styles.stepNumber}>1</span>
              <p className={styles.stepText}>
                Reply <strong>@threadmirror</strong> under any X (Twitter) thread
              </p>
            </div>
            <div className={styles.step}>
              <span className={styles.stepNumber}>2</span>
              <p className={styles.stepText}>
                Content is processed, summarized by AI, and uploaded to <strong>Filecoin</strong> network
              </p>
            </div>
            <div className={styles.step}>
              <span className={styles.stepNumber}>3</span>
              <p className={styles.stepText}>
                Access your decentralized archives forever through blockchain-verified links
              </p>
            </div>
          </div>
        </div>

        <div className={styles.cta}>
          <button 
            className={styles.loginButton}
            onClick={handleLogin}
          >
            Start mirroring
          </button>
          <p className={styles.ctaText}>
            Sign in to begin storing your content on the decentralized web with <strong>Filecoin</strong>.
          </p>
        </div>
      </div>
    </div>
  );
};

export default LoginPrompt; 