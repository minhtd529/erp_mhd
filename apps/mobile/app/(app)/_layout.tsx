import { Tabs } from 'expo-router';
import { View, Text, StyleSheet } from 'react-native';
import { colors } from '@/lib/theme';

function TabIcon({ emoji, label, focused }: { emoji: string; label: string; focused: boolean }) {
  return (
    <View style={styles.tabItem}>
      <Text style={[styles.tabEmoji, focused && styles.tabEmojiFocused]}>{emoji}</Text>
      <Text style={[styles.tabLabel, focused && styles.tabLabelFocused]}>{label}</Text>
    </View>
  );
}

export default function AppLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarStyle: styles.tabBar,
        tabBarShowLabel: false,
      }}
    >
      <Tabs.Screen
        name="dashboard"
        options={{
          tabBarIcon: ({ focused }) => <TabIcon emoji="🏠" label="Tổng quan" focused={focused} />,
        }}
      />
      <Tabs.Screen
        name="engagements"
        options={{
          tabBarIcon: ({ focused }) => <TabIcon emoji="📋" label="Hợp đồng" focused={focused} />,
        }}
      />
      <Tabs.Screen
        name="timesheet"
        options={{
          tabBarIcon: ({ focused }) => <TabIcon emoji="⏱" label="Chấm công" focused={focused} />,
        }}
      />
      <Tabs.Screen
        name="profile"
        options={{
          tabBarIcon: ({ focused }) => <TabIcon emoji="👤" label="Tài khoản" focused={focused} />,
        }}
      />
    </Tabs>
  );
}

const styles = StyleSheet.create({
  tabBar: {
    backgroundColor: colors.surface,
    borderTopColor: colors.border,
    height: 72,
    paddingBottom: 8,
  },
  tabItem: { alignItems: 'center', justifyContent: 'center', gap: 2, paddingTop: 6 },
  tabEmoji: { fontSize: 22, opacity: 0.5 },
  tabEmojiFocused: { opacity: 1 },
  tabLabel: { fontSize: 10, color: colors.textSecondary, fontWeight: '500' },
  tabLabelFocused: { color: colors.primary, fontWeight: '600' },
});
