import './UserProfile.css';
import { TweetUser } from '@client/index';

interface UserProfileProps {
    profile: TweetUser;
    sample?: boolean;
}

const UserProfileComponent = ({ profile, sample }: UserProfileProps) => {
    return (
        <div style={{ display: 'flex', alignItems: 'flex-start', marginBottom: '15px' }}>
            {/* 头像部分 */}
            <img
                src={profile.profileImageUrl}
                alt="Profile"
                style={{ width: '80px', height: '80px', borderRadius: '50%', marginRight: '15px' }}
            />
            {/* 右侧内容 */}
            <div style={{ flex: 1, textAlign: 'left' }}>
                {/* 第一行：名称 */}
                <h2 style={{ fontSize: '18px', fontWeight: 'bold', margin: '0 0 5px 0' }}>
                    <a href="javascript:void(0)">{profile.name}</a>
                </h2>
                {/* 第二行：screen_name */}
                <p style={{ fontSize: '14px', color: '#1da1f2', margin: '0 0 5px 0' }} >
                    <a href={`https://twitter.com/${profile.screenName}`} target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none', color: '#1da1f2' }}>
                        @{profile.screenName}
                    </a>
                </p>
                {/* 第三行：描述 */}
                <p style={{ fontSize: '14px', color: '#657786', margin: '0 0 10px 0' }}>
                    {profile.description}
                </p>
                {/* 最后一行：统计信息 */}
                {!sample && (
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', height: '20px' }}>
                        <div>
                            <span style={{ fontWeight: 'bold', color: 'gray', margin: '0 5px 0 0' }}>followers:</span>
                            <span style={{ color: '#b0b0b0' }}>{profile.followersCount}</span>

                            <span style={{ fontWeight: 'bold', color: 'gray', margin: '0 5px 0 10px' }}>friends:</span>
                            <span style={{ color: '#b0b0b0' }}>{profile.friendsCount}</span>

                            <span style={{ fontWeight: 'bold', color: 'gray', margin: '0 5px 0 10px' }}>mentions:</span>
                            <span style={{ color: '#b0b0b0' }}>{profile.statusesCount}</span>
                        </div>
                        <div>
                            <span style={{ fontWeight: 'bold', color: 'gray', margin: '0 5px 0 10px' }}>joined:</span>
                            <span style={{ color: '#b0b0b0' }}>{profile.createdAt.toDateString()}</span>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default UserProfileComponent;