"use client";

import { useEffect, useState } from "react";
import { User, LogOut, Settings, Users, Mail, Save, Trash2, PlusCircle, Edit3, Trash } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { useRouter } from "next/navigation";
import { useApi } from "@/lib/ApiContext";
import { toast } from "sonner";

interface UserInfo {
  id: number;
  email: string;
  nickname: string;
}

interface GroupInfo {
  id: number;
  name: string;
  owner_id: number;
  members: UserInfo[];
}

export default function Profile() {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [groups, setGroups] = useState<GroupInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [savingUser, setSavingUser] = useState(false);
  const [nickname, setNickname] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviting, setInviting] = useState(false);
  const [creatingGroup, setCreatingGroup] = useState(false);
  
  const router = useRouter();
  const { checkAuth } = useApi();

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      router.push("/login");
      return;
    }

    async function fetchData() {
      try {
        const userData = await apiRequest("/auth/me");
        setUser(userData);
        setNickname(userData.nickname);
        setEmail(userData.email);

        const groupData = await apiRequest("/api/groups");
        setGroups(groupData);
      } catch (err) {
        console.error("Failed to fetch profile data:", err);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [router]);

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    checkAuth();
    router.push("/login");
  };

  const handleCreateGroup = async () => {
    const groupName = prompt("新しいグループの名前を入力してください", "家族");
    if (!groupName) return;

    setCreatingGroup(true);
    try {
      await apiRequest("/api/groups", {
        method: "POST",
        body: JSON.stringify({ name: groupName }),
      });
      toast.success("グループを作成しました");
      // グループ一覧を再取得
      const groupData = await apiRequest("/api/groups");
      setGroups(groupData);
    } catch (err: any) {
      toast.error("グループ作成に失敗しました: " + err.message);
    } finally {
      setCreatingGroup(false);
    }
  };

  const handleUpdateGroup = async (groupId: number, currentName: string) => {
    const newName = prompt("グループの名前を変更します", currentName);
    if (!newName || newName === currentName) return;

    try {
      await apiRequest(`/api/groups/${groupId}`, {
        method: "PUT",
        body: JSON.stringify({ name: newName }),
      });
      toast.success("グループ名を変更しました");
      // 一覧を再取得
      const groupData = await apiRequest("/api/groups");
      setGroups(groupData);
    } catch (err: any) {
      toast.error("変更に失敗しました: " + err.message);
    }
  };

  const handleDeleteGroup = async (groupId: number, groupName: string) => {
    if (!confirm(`グループ「${groupName}」を削除しますか？\n(関連するすべてのレシートと精算履歴も削除されます)`)) return;

    try {
      await apiRequest(`/api/groups/${groupId}`, {
        method: "DELETE",
      });
      toast.success("グループを削除しました");
      // 一覧を再取得
      const groupData = await apiRequest("/api/groups");
      setGroups(groupData);
    } catch (err: any) {
      toast.error("削除に失敗しました: " + err.message);
    }
  };

  const handleUpdateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingUser(true);
    try {
      const updated = await apiRequest("/auth/me", {
        method: "PUT",
        body: JSON.stringify({ nickname, email, password: password || undefined }),
      });
      setUser(updated);
      setPassword("");
      toast.success("アカウント情報を更新しました");
    } catch (err: any) {
      toast.error("更新に失敗しました: " + err.message);
    } finally {
      setSavingUser(false);
    }
  };

  const handleInvite = async (groupId: number) => {
    if (!inviteEmail) return;
    setInviting(true);
    try {
      await apiRequest(`/api/groups/${groupId}/invite`, {
        method: "POST",
        body: JSON.stringify({ email: inviteEmail }),
      });
      setInviteEmail("");
      toast.success("メンバーを招待しました");
      // グループ情報を再取得
      const groupData = await apiRequest("/api/groups");
      setGroups(groupData);
    } catch (err: any) {
      toast.error("招待に失敗しました: " + err.message);
    } finally {
      setInviting(false);
    }
  };

  const handleRemoveMember = async (groupId: number, memberId: number) => {
    if (!confirm("本当にこのメンバーをグループから削除しますか？")) return;
    try {
      await apiRequest(`/api/groups/${groupId}/members/${memberId}`, {
        method: "DELETE",
      });
      toast.success("メンバーを削除しました");
      // グループ情報を再取得
      const groupData = await apiRequest("/api/groups");
      setGroups(groupData);
    } catch (err: any) {
      toast.error("削除に失敗しました: " + err.message);
    }
  };

  if (loading) return <div className="p-8 text-center text-gray-400">読み込み中...</div>;

  return (
    <div className="pb-10">
      <header className="p-4 border-b border-gray-100 bg-white sticky top-0 z-10 flex justify-between items-center">
        <h1 className="text-lg font-bold text-gray-800">設定</h1>
        <button onClick={handleLogout} className="text-red-500 flex items-center gap-1 text-sm font-bold">
          <LogOut size={18} />
          ログアウト
        </button>
      </header>

      <div className="p-6 space-y-8">
        {/* Account Info */}
        <section className="space-y-4">
          <div className="flex items-center gap-2 text-gray-400">
            <User size={18} />
            <h2 className="text-xs font-bold uppercase tracking-widest">アカウント情報</h2>
          </div>
          
          <form onSubmit={handleUpdateUser} className="bg-white border border-gray-100 rounded-3xl p-6 shadow-sm space-y-4">
            <div className="space-y-1">
              <label className="text-xs font-bold text-gray-500 ml-1">ニックネーム</label>
              <input 
                type="text" 
                className="w-full p-3 bg-gray-50 border border-gray-100 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-bold text-gray-500 ml-1">メールアドレス</label>
              <input 
                type="email" 
                className="w-full p-3 bg-gray-50 border border-gray-100 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-bold text-gray-500 ml-1">新しいパスワード (変更する場合のみ)</label>
              <input 
                type="password" 
                className="w-full p-3 bg-gray-50 border border-gray-100 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
              />
            </div>
            <button 
              type="submit" 
              disabled={savingUser}
              className="w-full py-3 bg-blue-600 text-white rounded-xl font-bold flex items-center justify-center gap-2 shadow-lg shadow-blue-100 active:scale-95 transition-all disabled:opacity-50"
            >
              <Save size={18} />
              保存する
            </button>
          </form>
        </section>

        {/* Group Management */}
        <section className="space-y-4">
          <div className="flex items-center gap-2 text-gray-400">
            <Users size={18} />
            <h2 className="text-xs font-bold uppercase tracking-widest">グループ管理</h2>
          </div>

          {groups.length === 0 ? (
            <div className="bg-gray-50 rounded-3xl p-8 text-center space-y-4">
              <p className="text-gray-500 text-sm">所属しているグループがありません</p>
              <button 
                onClick={handleCreateGroup}
                disabled={creatingGroup}
                className="text-blue-600 font-bold flex items-center gap-2 mx-auto disabled:opacity-50"
              >
                <PlusCircle size={20} />
                {creatingGroup ? "作成中..." : "グループを作成"}
              </button>
            </div>
          ) : (
            <>
              <div className="space-y-4">
                {groups.map(group => (
                  <div key={group.id} className="bg-white border border-gray-100 rounded-3xl shadow-sm overflow-hidden">
                    <div className="p-4 bg-gray-50 border-b border-gray-100 flex justify-between items-center">
                      <div className="flex items-center gap-2">
                        <h3 className="font-bold text-gray-800">{group.name}</h3>
                        {group.owner_id === user?.id && (
                          <button 
                            onClick={() => handleUpdateGroup(group.id, group.name)}
                            className="text-gray-400 hover:text-blue-500 transition-colors"
                          >
                            <Edit3 size={14} />
                          </button>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        {group.owner_id === user?.id && (
                          <button 
                            onClick={() => handleDeleteGroup(group.id, group.name)}
                            className="text-gray-300 hover:text-red-500 transition-colors mr-1"
                            title="グループを削除"
                          >
                            <Trash size={16} />
                          </button>
                        )}
                        <span className="text-[10px] font-bold bg-blue-100 text-blue-600 px-2 py-0.5 rounded">
                          ID: {group.id}
                        </span>
                      </div>
                    </div>
                    
                    <div className="p-4 space-y-4">
                      <div className="space-y-2">
                        <p className="text-[10px] font-bold text-gray-400 uppercase tracking-wider ml-1">メンバー</p>
                        <div className="divide-y divide-gray-50">
                          {group.members.map(member => (
                            <div key={member.id} className="py-3 flex justify-between items-center">
                              <div>
                                <p className="text-sm font-bold text-gray-800">{member.nickname}</p>
                                <p className="text-xs text-gray-500">{member.email}</p>
                              </div>
                              {group.owner_id === user?.id && member.id !== user?.id && (
                                <button 
                                  onClick={() => handleRemoveMember(group.id, member.id)}
                                  className="text-gray-300 hover:text-red-500 p-2"
                                >
                                  <Trash2 size={16} />
                                </button>
                              )}
                              {member.id === group.owner_id && (
                                <span className="text-[10px] font-bold text-orange-500 bg-orange-50 px-2 py-0.5 rounded">管理者</span>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>

                      {group.owner_id === user?.id && (
                        <div className="pt-4 border-t border-gray-50 space-y-2">
                          <p className="text-[10px] font-bold text-gray-400 uppercase tracking-wider ml-1">メンバーを招待</p>
                          <div className="flex gap-2">
                            <div className="relative flex-1">
                              <Mail size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                              <input 
                                type="email" 
                                placeholder="メールアドレス"
                                className="w-full p-2 pl-9 bg-gray-50 border border-gray-100 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
                                value={inviteEmail}
                                onChange={(e) => setInviteEmail(e.target.value)}
                              />
                            </div>
                            <button 
                              onClick={() => handleInvite(group.id)}
                              disabled={inviting || !inviteEmail}
                              className="bg-gray-900 text-white px-4 py-2 rounded-xl text-sm font-bold active:scale-95 transition-all disabled:opacity-50"
                            >
                              招待
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
              
              <button 
                onClick={handleCreateGroup}
                disabled={creatingGroup}
                className="w-full py-4 border-2 border-dashed border-gray-200 rounded-3xl text-gray-400 font-bold flex items-center justify-center gap-2 hover:bg-gray-50 active:scale-95 transition-all disabled:opacity-50"
              >
                <PlusCircle size={20} />
                {creatingGroup ? "作成中..." : "新しくグループを追加"}
              </button>
            </>
          )}
        </section>
      </div>
    </div>
  );
}
