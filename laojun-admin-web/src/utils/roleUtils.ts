import { User, Role } from '@/types';

/**
 * 用户角色管理工具函数
 */
export class RoleUtils {
  /**
   * 验证角色ID数组的有效性
   */
  static validateRoleIds(roleIds: string[]): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    if (!Array.isArray(roleIds)) {
      errors.push('角色ID必须是数组格式');
      return { isValid: false, errors };
    }
    
    if (roleIds.length === 0) {
      errors.push('至少需要选择一个角色');
      return { isValid: false, errors };
    }
    
    // 检查重复的角色ID
    const uniqueIds = new Set(roleIds);
    if (uniqueIds.size !== roleIds.length) {
      errors.push('存在重复的角色ID');
    }
    
    // 检查角色ID格式（UUID格式）
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    const invalidIds = roleIds.filter(id => !uuidRegex.test(id));
    if (invalidIds.length > 0) {
      errors.push(`无效的角色ID格式: ${invalidIds.join(', ')}`);
    }
    
    return { isValid: errors.length === 0, errors };
  }
  
  /**
   * 比较两个角色数组是否相同
   */
  static compareRoleArrays(roles1: string[], roles2: string[]): boolean {
    if (roles1.length !== roles2.length) return false;
    
    const sorted1 = [...roles1].sort();
    const sorted2 = [...roles2].sort();
    
    return JSON.stringify(sorted1) === JSON.stringify(sorted2);
  }
  
  /**
   * 从用户对象中提取角色ID数组
   */
  static extractRoleIds(user: User): string[] {
    return user.roles?.map(role => role.id) || [];
  }
  
  /**
   * 验证角色是否存在于可用角色列表中
   */
  static validateRolesExist(selectedRoleIds: string[], availableRoles: Role[]): { isValid: boolean; missingRoles: string[] } {
    const availableRoleIds = new Set(availableRoles.map(role => role.id));
    const missingRoles = selectedRoleIds.filter(id => !availableRoleIds.has(id));
    
    return {
      isValid: missingRoles.length === 0,
      missingRoles
    };
  }
  
  /**
   * 获取角色变更的差异
   */
  static getRoleChanges(oldRoleIds: string[], newRoleIds: string[]): {
    added: string[];
    removed: string[];
    unchanged: string[];
  } {
    const oldSet = new Set(oldRoleIds);
    const newSet = new Set(newRoleIds);
    
    const added = newRoleIds.filter(id => !oldSet.has(id));
    const removed = oldRoleIds.filter(id => !newSet.has(id));
    const unchanged = oldRoleIds.filter(id => newSet.has(id));
    
    return { added, removed, unchanged };
  }
  
  /**
   * 格式化角色变更信息为用户友好的文本
   */
  static formatRoleChanges(changes: ReturnType<typeof RoleUtils.getRoleChanges>, availableRoles: Role[]): string {
    const roleMap = new Map(availableRoles.map(role => [role.id, role.name]));
    const messages: string[] = [];
    
    if (changes.added.length > 0) {
      const addedNames = changes.added.map(id => roleMap.get(id) || id).join(', ');
      messages.push(`新增角色: ${addedNames}`);
    }
    
    if (changes.removed.length > 0) {
      const removedNames = changes.removed.map(id => roleMap.get(id) || id).join(', ');
      messages.push(`移除角色: ${removedNames}`);
    }
    
    if (changes.unchanged.length > 0) {
      const unchangedNames = changes.unchanged.map(id => roleMap.get(id) || id).join(', ');
      messages.push(`保持不变: ${unchangedNames}`);
    }
    
    return messages.join('; ');
  }
  
  /**
   * 检查用户是否具有系统角色
   */
  static hasSystemRole(user: User): boolean {
    return user.roles?.some(role => role.isSystem) || false;
  }
  
  /**
   * 过滤出系统角色
   */
  static filterSystemRoles(roles: Role[]): Role[] {
    return roles.filter(role => role.isSystem);
  }
  
  /**
   * 过滤出普通角色
   */
  static filterNormalRoles(roles: Role[]): Role[] {
    return roles.filter(role => !role.isSystem);
  }
}