import type { UserData } from './utils/user-data'

function ProfileCard({ userData }: { userData: UserData }) {
  return (
   <div
     className={`
       grid grid-cols-1 md:grid-cols-[max-content_1fr]
       bg-tn-d-black text-tn-d-fg text-2xl
       border-2 border-tn-d-fg rounded-2xl my-2 pb-0 md:pb-4 p-4 md:gap-y-2
     `}
   >
     <ProfileItem title="Username" value={
       userData ? userData.username : ""
     } />
     <ProfileItem title="Email" value={userData.email} />
     <ProfileItem title="Role" value={userData.role} />
   </div>
  );
}

function ProfileItem({ title, value }: { title: string, value: string }) {
  return (
    <>
      <span className="font-bold">{ title }:</span>
      <div
        className={`
          overflow-x-auto whitespace-nowrap
          pb-4 md:pb-0 md:ps-4
        `}
      >
        <span>{ value }</span>
      </div>
    </>
  );
}

export default ProfileCard;
