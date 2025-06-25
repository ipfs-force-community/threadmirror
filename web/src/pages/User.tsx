import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import UserProfileComponent from '@components/common/UserProfile';
import ThreadList from '@components/thread/ThreadList';
import { UserProfile, UserTwitterResponse } from '@src/types';
import { useApiService } from '@services/api';
import { convTweetData2Thread } from '../types/format';
import styles from './User.module.css';

const User = () => {
  const { id } = useParams<{ id: string, name: string }>();
  const [userTwitters, setUserTwitters] = useState<UserTwitterResponse | null>(null);
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const { fetchUserTwitter } = useApiService();
  
  useEffect(() => {
    const loadData = async () => {
      try {
        // if (!id) throw new Error('No ID provided');
        const detailData = await fetchUserTwitter(id || '', 20, 0);
        setUserTwitters(detailData);
        setUserProfile(detailData?.auth || null);
      } catch (err) {
        console.error(err);
      } finally {
        // setLoading(false);
      }
    };
    loadData();
  }, [id, fetchUserTwitter]);

  return (
    <div className={styles.user_page}>
      {!userProfile ? <div>作者信息缺失</div> :
        (
          <div>
            <UserProfileComponent profile={userProfile} />
            <ThreadList threads={userTwitters?.msg?.map(tweetData => convTweetData2Thread(tweetData)) || []} />
          </div>
        )
      }
    </div>
  );
};

export default User;